package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"context"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	"github.com/jackdoe/baxx/api/file"
	"github.com/jackdoe/baxx/api/helpers"
	"github.com/jackdoe/baxx/monitoring"

	al "github.com/jackdoe/baxx/api/action_log"
	"github.com/jackdoe/baxx/api/user"
	"github.com/jackdoe/baxx/help"
	"github.com/jackdoe/baxx/message"
)

type server struct {
	db         *gorm.DB
	store      *file.Store
	r          *gin.Engine
	authorized *gin.RouterGroup
}

func (s *server) registerHelp(protected bool, helpobj help.HelpObject, path ...string) {
	for _, p := range path {
		if protected {
			s.authorized.GET("/help"+p, func(c *gin.Context) {
				c.String(http.StatusOK, help.Render(helpobj))
			})
		} else {
			s.r.GET("/help"+p, func(c *gin.Context) {
				c.String(http.StatusOK, help.Render(helpobj))
			})
		}
	}
}

func (s *server) getViewTokenLoggedOrNot(c *gin.Context) (*file.Token, *user.User, error) {
	token := c.Param("token")
	var t *file.Token
	var u *user.User
	var err error
	x, isLoggedIn := c.Get("user")
	if isLoggedIn {
		u = x.(*user.User)
		t, err = helpers.FindTokenForUser(s.db, token, u)
		if err != nil {
			return nil, nil, err
		}
	} else {
		t, u, err = helpers.FindTokenAndUser(s.db, token)
		if err != nil {
			return nil, nil, err
		}
		writing := c.Request.Method == "POST" || c.Request.Method == "PUT"
		if t.WriteOnly && !writing {
			return nil, nil, fmt.Errorf("write only token, use basic auth curl -X%s-u your.email https://baxx.dev/{io,ls}/$TOKEN/*path", c.Request.Method)
		}
	}

	if u.EmailVerified == nil {
		err := errors.New("email not verified yet")
		return nil, nil, err
	}

	if !u.Paid() {
		err := fmt.Errorf("payment not received yet or subscription is cancelled, go to https://baxx.dev/sub/%s or if you already did, contact me at jack@baxx.dev", u.PaymentID)
		return nil, nil, err
	}

	return t, u, nil
}

func setupAPI(db *gorm.DB, bind string) {
	j := os.Getenv("BAXX_JUDOC_URL")
	if j == "" {
		j = "http://localhost:9122"
	}
	store, err := file.NewStore(j)

	if err != nil {
		log.Panic(err)
	}

	r := gin.Default()
	r.Use(SlackRecovery())
	r.Use(func(c *gin.Context) {
		su, pass := BasicAuthDecode(c)
		if su != "" {
			u, err := user.FindUser(db, su, pass)
			if err == nil {
				if u.MatchPassword(pass) {
					c.Set("user", u)
					if c.Request.RequestURI != "/protected/status" {
						al.Log(db, u.ID, c.Request.Method, c.Request.RequestURI, c.Request)
					}
				}
			}
		}
	})

	authorized := r.Group("/protected") // FIXME rename that? its too long /p/ should be fine
	authorized.Use(func(c *gin.Context) {
		_, loggedIn := c.Get("user")
		if !loggedIn {
			c.Header("WWW-Authenticate", "Authorization Required")
			c.String(401, `{"error": "Not Authorized (auth required, or wrong password)"}`)
			c.Abort()
		}
	})

	r.GET("/digitalocean", func(c *gin.Context) {
		c.String(200, "ok")
	})

	r.GET("/join/groups", func(c *gin.Context) {
		url := "https://groups.google.com/forum/#!forum/baxx-users"
		c.Redirect(http.StatusFound, url)
	})

	r.GET("/join/slack", func(c *gin.Context) {
		url := "https://join.slack.com/t/baxxdev/shared_invite/enQtNTkwODg2ODYyOTE5LTVhMDkwYzVlMTIxNWJkYzMzZjI1ZGIxYjdjYTZhNzkyZTRiOTlmNWI2ZDU5MDZiOWUwZjAzNTU1MDViYmJkODM"
		c.Redirect(http.StatusFound, url)
	})

	r.GET("/", func(c *gin.Context) {
		c.String(200, "sign up: ssh register@ui.baxx.dev\nhelp: curl https://baxx.dev/help\nslack: https://baxx.dev/join/slack\ngoogle groups: https://baxx.dev/join/groups")
	})

	r.GET("/stat", func(c *gin.Context) {
		s, err := monitoring.GetStats(db, monitoring.Hostname(), 60)
		if err != nil {
			warnErr(c, err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		b := new(bytes.Buffer)
		height := 15
		for _, st := range s {
			st.ASCII(b, height)
		}
		c.String(200, b.String())
	})
	srv := &server{db: db, r: r, store: store, authorized: authorized}
	setupIO(srv)
	setupACC(srv)
	setupSYNC(srv)
	mux := &http.Server{
		Addr:    bind,
		Handler: r,
	}
	go func() {
		if err := mux.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Panicf("listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := mux.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

func main() {
	message.MustHavePanic()
	defer message.SlackPanic("main api")

	var pbind = flag.String("bind", ":9123", "bind")
	var pdebug = flag.Bool("debug", false, "debug")
	var prelease = flag.Bool("release", false, "release")
	flag.Parse()

	CONFIG.MaxTokens = 100

	if *prelease {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := gorm.Open("postgres", os.Getenv("BAXX_POSTGRES"))
	if err != nil {
		log.Panic(err)
	}
	db.LogMode(*pdebug)
	defer db.Close()

	setupAPI(db, *pbind)
}
