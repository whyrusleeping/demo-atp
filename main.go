package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	cli "github.com/urfave/cli/v2"
	"github.com/whyrusleeping/demo-atp/records"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Server struct {
	db *gorm.DB

	cursor   int64
	cursorLk sync.Mutex

	relayHost string

	userLk    sync.Mutex
	userCache *lru.Cache[string, *User]

	directory identity.Directory
}

type User struct {
	gorm.Model
	Did    string `gorm:"uniqueIndex"`
	Handle string

	Lk sync.Mutex `gorm:"-"`
}

type Comment struct {
	gorm.Model

	Author  uint `gorm:"uniqueIndex:idx_author_rkey"`
	Profile uint
	Created time.Time
	Rkey    string `gorm:"uniqueIndex:idx_author_rkey"`
	Text    string
}

type UserProfile struct {
	gorm.Model
	Author uint `gorm:"uniqueIndex"`
	Data   string
}

func main() {

	app := cli.NewApp()

	app.Commands = []*cli.Command{
		createProfileCmd,
		createCommentCmd,
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "relay-host",
			Value: "wss://bsky.network",
		},
	}
	app.Action = func(cctx *cli.Context) error {
		db, err := gorm.Open(sqlite.Open("demo.db"))
		if err != nil {
			return err
		}

		db.AutoMigrate(&LastSeq{})
		db.AutoMigrate(&User{})
		db.AutoMigrate(&Comment{})
		db.AutoMigrate(&UserProfile{})

		uc, _ := lru.New[string, *User](100_000)
		s := &Server{
			db:        db,
			relayHost: cctx.String("relay-host"),
			userCache: uc,
			directory: identity.DefaultDirectory(),
		}

		go s.Run(context.TODO())

		return s.Serve(":9987")
	}

	app.RunAndExitOnError()
}

func (s *Server) Serve(addr string) error {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))

	e.GET("/getCommentsForPage/:user", s.handleGetComments)
	e.GET("/getProfileData/:user", s.handleGetProfileData)

	return e.Start(addr)
}

func (s *Server) uidToDid(uid uint) (string, error) {
	var u User
	if err := s.db.Find(&u, "id = ?", uid).Error; err != nil {
		return "", err
	}

	if u.ID == 0 {
		return "", fmt.Errorf("unknown user")
	}

	return u.Did, nil
}

func (s *Server) getUserByIdent(ctx context.Context, ident string) (*User, error) {
	val, err := syntax.ParseAtIdentifier(ident)
	if err != nil {
		return nil, err
	}

	if !val.IsDID() {
		var uh User
		if err := s.db.Find(&uh, "handle = ?", ident).Error; err != nil {
			return nil, err
		}

		if uh.ID != 0 {
			return &uh, nil
		}

		h, err := val.AsHandle()
		if err != nil {
			return nil, err
		}

		resp, err := s.directory.LookupHandle(ctx, h)
		if err != nil {
			return nil, err
		}

		if err := s.db.Model(User{}).Where("did = ?", resp.DID).Update("handle", resp.Handle).Error; err != nil {
			return nil, err
		}

		return s.getOrCreateUser(ctx, resp.DID.String())
	}

	return s.getOrCreateUser(ctx, ident)
}

type apiComment struct {
	Author  string    `json:"author"`
	Created time.Time `json:"created"`
	Text    string    `json:"text"`
}

func (s *Server) handleGetComments(cctx echo.Context) error {
	user := cctx.Param("user")
	ctx := context.TODO()

	u, err := s.getUserByIdent(ctx, user)
	if err != nil {
		return err
	}

	var dbc []Comment
	if err := s.db.Find(&dbc, "profile = ?", u.ID).Error; err != nil {
		return err
	}

	var comments []apiComment
	for _, c := range dbc {
		adid, err := s.uidToDid(c.Author)
		if err != nil {
			return err
		}

		comments = append(comments, apiComment{
			Author:  adid,
			Created: c.Created,
			Text:    c.Text,
		})
	}
	return cctx.JSON(200, comments)
}

type apiProfile struct {
	Handle string   `json:"handle"`
	Text   string   `json:"text"`
	Links  []string `json:"links"`
}

func (s *Server) handleGetProfileData(cctx echo.Context) error {
	user := cctx.Param("user")
	ctx := context.TODO()

	u, err := s.getUserByIdent(ctx, user)
	if err != nil {
		return err
	}

	var prof UserProfile
	if err := s.db.Find(&prof, "author = ?", u.ID).Error; err != nil {
		return err
	}

	var out apiProfile
	if err := json.Unmarshal([]byte(prof.Data), &out); err != nil {
		return err
	}

	out.Handle = u.Handle

	return cctx.JSON(200, out)
}

func (s *Server) handleCreateComment(ctx context.Context, u *User, path string, rec *records.Comment) error {
	pathparts := strings.Split(path, "/")

	log.Println("Handling create comment: ", u.Handle, u.Did)
	profu, err := s.getOrCreateUser(ctx, rec.Profile)
	if err != nil {
		return err
	}

	var p UserProfile
	if err := s.db.Find(&p, "author = ?", profu.ID).Error; err != nil {
		return err
	}

	if p.ID == 0 {
		return fmt.Errorf("comment for non-existent profile")
	}

	t := time.Unix(int64(rec.CreatedAt), 0)

	cmt := Comment{
		Author:  u.ID,
		Profile: profu.ID,
		Created: t,
		Text:    rec.Text,
		Rkey:    pathparts[1],
	}

	if err := s.db.Create(&cmt).Error; err != nil {
		return err
	}

	return nil
}

func (s *Server) handleCreateProfile(ctx context.Context, u *User, path string, rec *records.Profile) error {
	log.Println("Handling create profile: ", u.Handle, u.Did)
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	profile := UserProfile{
		Author: u.ID,
		Data:   string(b),
	}

	if err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "author"}},
		UpdateAll: true,
	}).Create(&profile).Error; err != nil {
		return err
	}

	return nil
}
