package ginserver

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"notes/internal/notes/app"
	"notes/internal/notes/server/ginserver/middlewares"
	"notes/internal/pkg/config"
	"notes/internal/pkg/logger"
	"notes/internal/pkg/models"

	"github.com/gin-gonic/gin"
)

type Server struct {
	a    app.App
	cfg  config.Server
	srv  *http.Server
	logg logger.Logger
	e    *gin.Engine
}

func New(a app.App, cfg config.Server, logg logger.Logger) *Server {
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
	notes, err := s.a.GetNotes(ctx)
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
		if errors.Is(err, sql.ErrNoRows) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(200, note)
}

func (s *Server) CreateNote(c *gin.Context) {
	var n models.Note
	c.BindJSON(&n)

	ctx := context.Background()
	if err := s.a.CreateNote(ctx, n); err != nil {
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
	c.BindJSON(&n)

	ctx := context.Background()
	if err := s.a.UpdateNote(ctx, n); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Header("Content-Type", "application/json")
	c.Status(http.StatusNoContent)
}
