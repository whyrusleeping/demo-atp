package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/autoscaling"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/repomgr"
	"github.com/gorilla/websocket"
	"github.com/ipfs/go-cid"
	"github.com/whyrusleeping/demo-atp/records"
)

func init() {
	lexutil.RegisterType("world.bsky.demo.profile", &records.Profile{})
	lexutil.RegisterType("world.bsky.demo.comment", &records.Comment{})
}

// NB: This scaffolding mostly copied from whyrusleeping/algoz

type LastSeq struct {
	ID  uint `gorm:"primarykey"`
	Seq int64
}

func (s *Server) loadCursor() (int64, error) {
	var lastSeq LastSeq
	if err := s.db.Find(&lastSeq).Error; err != nil {
		return 0, err
	}

	if lastSeq.ID == 0 {
		return 0, s.db.Create(&lastSeq).Error
	}

	return lastSeq.Seq, nil
}

func (s *Server) getCursor() int64 {
	s.cursorLk.Lock()
	defer s.cursorLk.Unlock()
	return s.cursor
}

func (s *Server) updateLastCursor(curs int64) error {
	s.cursorLk.Lock()
	if curs < s.cursor {
		s.cursorLk.Unlock()
		return nil
	}
	s.cursor = curs
	s.cursorLk.Unlock()

	if curs%200 == 0 {
		if err := s.db.Model(LastSeq{}).Where("id = 1").Update("seq", curs).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) areWeInterested(evt *atproto.SyncSubscribeRepos_Commit) bool {
	for _, op := range evt.Ops {
		if strings.HasPrefix(op.Path, "world.bsky.demo") {
			return true
		}
	}

	return false
}

func (s *Server) Run(ctx context.Context) error {
	log.Println("Starting up...")

	loadedCursor, err := s.loadCursor()
	if err != nil {
		return fmt.Errorf("get last cursor: %w", err)
	}

	s.cursor = loadedCursor

	handleFunc := func(ctx context.Context, xe *events.XRPCStreamEvent) error {
		switch {
		case xe.RepoCommit != nil:
			evt := xe.RepoCommit

			if err := s.updateLastCursor(evt.Seq); err != nil {
				log.Printf("Failed to update cursor: %s", err)
			}

			if !s.areWeInterested(evt) {
				return nil
			}

			log.Println()
			if evt.TooBig && evt.Prev != nil {
				log.Printf("skipping non-genesis too big events for now: %d", evt.Seq)
				return nil
			}

			if evt.TooBig {
				return nil
				if err := s.processTooBigCommit(ctx, evt); err != nil {
					log.Printf("failed to process tooBig event: %s", err)
					return nil
				}

				return nil
			}

			r, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(evt.Blocks))
			if err != nil {
				log.Printf("reading repo from car (seq: %d, len: %d): %s", evt.Seq, len(evt.Blocks), err)
				return nil
			}

			for _, op := range evt.Ops {
				ek := repomgr.EventKind(op.Action)
				switch ek {
				case repomgr.EvtKindCreateRecord, repomgr.EvtKindUpdateRecord:
					rc, rec, err := r.GetRecord(ctx, op.Path)
					if err != nil {
						e := fmt.Errorf("getting record %s (%s) within seq %d for %s: %w", op.Path, *op.Cid, evt.Seq, evt.Repo, err)
						log.Printf(e.Error())
						return nil
					}

					if lexutil.LexLink(rc) != *op.Cid {
						log.Printf("mismatch in record and op cid: %s != %s", rc, *op.Cid)
						return nil
					}

					if err := s.handleOp(ctx, ek, evt.Seq, op.Path, evt.Repo, &rc, rec); err != nil {
						log.Printf("failed to handle op: %s", err)
						return nil
					}

				case repomgr.EvtKindDeleteRecord:
					if err := s.handleOp(ctx, ek, evt.Seq, op.Path, evt.Repo, nil, nil); err != nil {
						log.Printf("failed to handle delete: %s", err)
						return nil
					}
				}
			}

			return nil
		case xe.RepoHandle != nil:
			evt := xe.RepoHandle
			if err := s.updateUserHandle(ctx, evt.Did, evt.Handle); err != nil {
				log.Printf("failed to update user handle: %s", err)
			}
			return nil
		default:
			return nil
		}
	}

	var backoff time.Duration
	for {
		d := websocket.DefaultDialer
		con, _, err := d.Dial(fmt.Sprintf("%s/xrpc/com.atproto.sync.subscribeRepos?cursor=%d", s.relayHost, s.getCursor()), http.Header{})
		if err != nil {
			log.Printf("failed to dial: %s", err)
			time.Sleep(backoff)

			if backoff < time.Minute {
				backoff = (backoff * 2) + time.Second
			}
			continue
		}

		backoff = 0

		opts := autoscaling.DefaultAutoscaleSettings()
		opts.Concurrency = 20
		opts.MaxConcurrency = 100
		sched := autoscaling.NewScheduler(opts, "", handleFunc)
		if err := events.HandleRepoStream(ctx, con, sched); err != nil {
			log.Printf("stream processing error: %s", err)
		}
	}
}

// handleOp receives every incoming repo event and is where indexing logic lives
func (s *Server) handleOp(ctx context.Context, op repomgr.EventKind, seq int64, path string, did string, rcid *cid.Cid, rec any) error {
	col := strings.Split(path, "/")[0]
	if op == repomgr.EvtKindCreateRecord || op == repomgr.EvtKindUpdateRecord {
		log.Printf("handling event(%d): %s - %s", seq, did, path)
		u, err := s.getOrCreateUser(ctx, did)
		if err != nil {
			return fmt.Errorf("checking user: %w", err)
		}

		_ = col
		_ = u
		switch rec := rec.(type) {
		case *records.Comment:
			return s.handleCreateComment(ctx, u, path, rec)
		case *records.Profile:
			return s.handleCreateProfile(ctx, u, path, rec)
		default:
		}

	} else if op == repomgr.EvtKindDeleteRecord {
		u, err := s.getOrCreateUser(ctx, did)
		if err != nil {
			return err
		}

		_ = u

		parts := strings.Split(path, "/")
		// Not handling like/repost deletes because it requires individually tracking *every* single like
		switch parts[0] {
		// TODO: handle profile deletes, its an edge case, but worth doing still
		case "app.bsky.feed.post":
			/*
				if err := s.deletePost(ctx, u, path); err != nil {
					return err
				}
			*/
		}
	}

	if err := s.updateLastCursor(seq); err != nil {
		log.Printf("Failed to update cursor: %s", err)
	}

	return nil
}

func (s *Server) processTooBigCommit(ctx context.Context, evt *atproto.SyncSubscribeRepos_Commit) error {
	/*

		repodata, err := atproto.SyncGetRepo(ctx, s.bgsxrpc, evt.Repo, "")
		if err != nil {
			return err
		}

		r, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(repodata))
		if err != nil {
			return err
		}

		u, err := s.getOrCreateUser(ctx, evt.Repo)
		if err != nil {
			return err
		}

		return r.ForEach(ctx, "", func(k string, v cid.Cid) error {
			rcid, rec, err := r.GetRecord(ctx, k)
			if err != nil {
				log.Printf("failed to get record from repo checkout: %s", err)
				return nil
			}

			return s.handleOp(ctx, repomgr.EvtKindCreateRecord, evt.Seq, k, u.Did, &rcid, rec)
		})
	*/
	return nil
}

func (s *Server) getOrCreateUser(ctx context.Context, did string) (*User, error) {
	s.userLk.Lock()
	cu, ok := s.userCache.Get(did)
	if ok {
		s.userLk.Unlock()
		cu.Lk.Lock()
		cu.Lk.Unlock()
		if cu.ID == 0 {
			return nil, fmt.Errorf("user creation failed")
		}

		return cu, nil
	}

	var u User
	s.userCache.Add(did, &u)

	u.Lk.Lock()
	defer u.Lk.Unlock()
	s.userLk.Unlock()

	if err := s.db.Find(&u, "did = ?", did).Error; err != nil {
		return nil, err
	}
	if u.ID == 0 {
		// TODO: figure out peoples handles
		/*
			h, err := s.handleFromDid(ctx, did)
			if err != nil {
				log.Printf("failed to resolve did to handle", "did", did, "err", err)
			} else {
				u.Handle = h
			}
		*/

		u.Did = did
		if err := s.db.Create(&u).Error; err != nil {
			s.userCache.Remove(did)

			return nil, err
		}
	}

	return &u, nil
}

func (s *Server) handleFromDid(ctx context.Context, did string) (string, error) {
	resp, err := s.directory.LookupDID(ctx, syntax.DID(did))
	if err != nil {
		return "", err
	}

	return resp.Handle.String(), nil
}

func (s *Server) updateUserHandle(ctx context.Context, did string, handle string) error {
	u, err := s.getOrCreateUser(ctx, did)
	if err != nil {
		return err
	}

	return s.db.Model(&User{}).Where("id = ?", u.ID).Update("handle", handle).Error
}
