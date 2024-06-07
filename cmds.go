package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/urfave/cli/v2"
	"github.com/whyrusleeping/demo-atp/records"
)

var createProfileCmd = &cli.Command{
	Name: "create-profile",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "pds-host",
			Value: "https://bsky.social",
		},
		&cli.StringFlag{
			Name:     "handle",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "password",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {

		text := cctx.Args().First()
		links := cctx.Args().Slice()[1:]

		profile := &records.Profile{
			CreatedAt: time.Now().Unix(),
			Links:     links,
			Text:      text,
		}

		rkey := "self"

		xrpcc := &xrpc.Client{
			Host: cctx.String("pds-host"),
		}

		resp, err := atproto.ServerCreateSession(context.TODO(), xrpcc, &atproto.ServerCreateSession_Input{
			Identifier: cctx.String("handle"),
			Password:   cctx.String("password"),
		})
		if err != nil {
			return err
		}

		xrpcc.Auth = &xrpc.AuthInfo{
			AccessJwt:  resp.AccessJwt,
			RefreshJwt: resp.RefreshJwt,
			Handle:     resp.Handle,
			Did:        resp.Did,
		}

		validate := false
		out, err := atproto.RepoPutRecord(context.TODO(), xrpcc, &atproto.RepoPutRecord_Input{
			Collection: "world.bsky.demo.profile",
			Record:     &util.LexiconTypeDecoder{Val: profile},
			Repo:       resp.Did,
			Rkey:       rkey,
			Validate:   &validate,
		})
		if err != nil {
			return err
		}

		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(b))

		return nil
	},
}

var createCommentCmd = &cli.Command{
	Name: "create-comment",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "pds-host",
			Value: "https://bsky.social",
		},
		&cli.StringFlag{
			Name:     "handle",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "password",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {

		profile := cctx.Args().Get(0)
		text := cctx.Args().Get(1)

		xrpcc := &xrpc.Client{
			Host: cctx.String("pds-host"),
		}

		ctx := context.TODO()

		resp, err := atproto.ServerCreateSession(ctx, xrpcc, &atproto.ServerCreateSession_Input{
			Identifier: cctx.String("handle"),
			Password:   cctx.String("password"),
		})
		if err != nil {
			return err
		}

		if !strings.HasPrefix(profile, "did:") {
			dir := identity.DefaultDirectory()
			resp, err := dir.LookupHandle(ctx, syntax.Handle(profile))
			if err != nil {
				return err
			}

			profile = resp.DID.String()
		}

		xrpcc.Auth = &xrpc.AuthInfo{
			AccessJwt:  resp.AccessJwt,
			RefreshJwt: resp.RefreshJwt,
			Handle:     resp.Handle,
			Did:        resp.Did,
		}

		comment := &records.Comment{
			CreatedAt: time.Now().Unix(),
			Profile:   profile,
			Text:      text,
		}

		validate := false
		out, err := atproto.RepoCreateRecord(context.TODO(), xrpcc, &atproto.RepoCreateRecord_Input{
			Collection: "world.bsky.demo.comment",
			Record:     &util.LexiconTypeDecoder{Val: comment},
			Repo:       resp.Did,
			Validate:   &validate,
		})
		if err != nil {
			return err
		}

		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(b))

		return nil
	},
}
