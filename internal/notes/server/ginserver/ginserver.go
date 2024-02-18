package ginserver

import (
	"context"
	"errors"
	"net/http"
	"notes/internal/notes/server"
	"notes/internal/notes/server/ginserver/middlewares"
	"notes/internal/notes/storage"
	"notes/internal/pkg/config"
	"notes/internal/pkg/logger"
	"notes/internal/pkg/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	a    server.App
	cfg  config.Server
	srv  *http.Server
	logg logger.Logger
	e    *gin.Engine
}

func New(a server.App, cfg config.Server, logg logger.Logger) *Server {
	s := &Server{
		a:   a,
		cfg: cfg,
		srv: &http.Server{
			Addr:              cfg.Host + cfg.Port,
			ReadHeaderTimeout: time.Duration(cfg.ShutDownTimeout) * time.Second,
		},
		logg: logg,
	}

	s.Register()
	return s
}

// this function is made for testing server.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.srv.Handler.ServeHTTP(w, r)
}

func (s *Server) Start(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	default:
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
		// return s.e.Run(s.srv.Addr)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(s.cfg.ShutDownTimeout))
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Server) Register() {
	e := gin.Default()

	e.Use(gin.Recovery())
	e.Use(middlewares.LoggingMiddleware(s.logg))

	notes := e.Group("/notes")
	notes.GET("/", s.GetNotes)
	notes.GET("/:id", s.GetNote)
	notes.PUT("/", s.CreateNote)
	notes.DELETE("/", s.DeleteNote)
	notes.PATCH("/", s.UpdateNote)
	s.e = e
	s.srv.Handler = e
}

func (s *Server) GetNotes(c *gin.Context) {
	ctx := context.Background()
	dur, ok := c.GetQuery("interval")
	var interval time.Duration
	if ok {
		var err error
		interval, err = time.ParseDuration(dur)
		if err != nil {
			if err := c.AbortWithError(http.StatusBadRequest, err); err != nil {
				return
			}
			return
		}
	}

	notes, err := s.a.GetNotes(ctx, interval)
	if err != nil {
		if err := c.AbortWithError(http.StatusInternalServerError, err); err.Err != nil {
			return
		}
		return
	}
	c.JSON(http.StatusOK, notes)
}

func (s *Server) GetNote(c *gin.Context) {
	ctx := context.Background()
	noteID := c.Param("id")
	id, err := strconv.ParseUint(noteID, 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	note, err := s.a.GetNote(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	s.logg.Debugf("get note debug: note %v\n", note)
	c.JSON(200, note)
}

func (s *Server) CreateNote(c *gin.Context) {
	var n models.Note
	if err := c.BindJSON(&n); err != nil {
		s.logg.Debugf("create note debug: error: %v\n", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	s.logg.Debugf("create note debug: note: %v\n", n)
	ctx := context.Background()
	if err := s.a.CreateNote(ctx, n); err != nil {
		s.logg.Debugf("error: %v\n", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Header("Content-Type", "application/json")
	c.Status(http.StatusCreated)
}

func (s *Server) DeleteNote(c *gin.Context) {
	idS, ok := c.GetQuery("id")
	if !ok {
		c.AbortWithStatus(http.StatusNotAcceptable)
		return
	}
	id, err := strconv.ParseUint(idS, 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx := context.Background()
	if err := s.a.DeleteNote(ctx, id); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Header("Content-Type", "application/json")
	c.Status(http.StatusNoContent)
}

func (s *Server) UpdateNote(c *gin.Context) {
	var n models.Note
	if err := c.BindJSON(&n); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	s.logg.Debugf("update note debug: note: %v\n", n)
	s.logg.Debugf("refreshed header: %s\n", c.GetHeader("Refreshed"))
	ctx := context.Background()
	switch {
	case c.GetHeader("Refreshed") != "":
		if c.GetHeader("Refreshed") != "true" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		s.logg.Debugf("refresh note debug: note: %v\n", n)

		if err := s.a.RefreshNote(ctx, n); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	default:
		if err := s.a.UpdateNote(ctx, n); err != nil {
			switch {
			case errors.Is(err, storage.ErrNotFound):
				c.AbortWithStatus(http.StatusNotFound)
				return
			case errors.Is(err, storage.ErrNotEnoughArguments):
				c.AbortWithError(http.StatusBadRequest, err)
				return
			}
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	c.Header("Content-Type", "application/json")
	c.Status(http.StatusNoContent)
}
