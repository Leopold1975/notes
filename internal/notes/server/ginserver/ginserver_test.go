package ginserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"notes/internal/notes/app"
	"notes/internal/notes/server/ginserver"
	"notes/internal/notes/storage"
	"notes/internal/pkg/config"
	"notes/internal/pkg/logger"
	"notes/internal/pkg/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var nilError error

func TestBasic(t *testing.T) {
	mockStr := new(storage.MockStorage)

	logg, err := logger.New(logger.EnvLocal)
	require.NoError(t, err)

	mockApp := app.NewApp(mockStr)

	cfg := config.Server{Host: "test", Port: ":80", ShutDownTimeout: 5}

	serv := ginserver.New(mockApp, cfg, logg)

	ctx := context.Background()

	t.Parallel()
	t.Run("Test Get Notes", func(t *testing.T) {
		w := httptest.NewRecorder()

		exp := []models.Note{}
		var td time.Duration
		mockStr.On("GetNotes", ctx, td).Return(exp, nilError)

		req, err := http.NewRequestWithContext(ctx, "GET", "/notes/", nil)
		assert.NoError(t, err)

		serv.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})

	t.Run("Test Get Notes With interval", func(t *testing.T) {
		w := httptest.NewRecorder()

		exp := []models.Note{}
		mockStr.On("GetNotes", ctx, time.Minute*5).Return(exp, nilError)

		req, err := http.NewRequestWithContext(ctx, "GET", "/notes/?interval=5m", nil)
		assert.NoError(t, err)

		serv.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})

	t.Run("Test Create Note", func(t *testing.T) {
		w := httptest.NewRecorder()
		tm, err := time.Parse("02.01.2006 15:04", "14.01.2024 11:03")
		assert.NoError(t, err)
		note := models.Note{
			Title:       "test",
			Description: "test",
			DateAdded:   tm,
			DateNotify:  tm,
		}

		mockStr.On("CreateNote", ctx, note).Return(nilError)

		b, err := json.Marshal(note)
		assert.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, "PUT", "/notes/", bytes.NewReader(b))
		assert.NoError(t, err)

		serv.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Result().StatusCode)
		assert.NoError(t, err)
	})

	t.Run("Test Get Note", func(t *testing.T) {
		w := httptest.NewRecorder()
		tm, err := time.Parse("02.01.2006 15:04", "14.01.2024 11:03")
		assert.NoError(t, err)
		note := models.Note{
			ID:          1,
			Title:       "test",
			Description: "test",
			DateAdded:   tm,
			DateNotify:  tm,
		}

		mockStr.On("GetNote", ctx, note.ID).Return(note, nilError)

		req, err := http.NewRequestWithContext(ctx, "GET", "/notes/1", nil)
		assert.NoError(t, err)

		serv.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.NoError(t, err)

		noteRes := models.Note{}
		err = json.Unmarshal(w.Body.Bytes(), &noteRes)
		assert.NoError(t, err)
		assert.Equal(t, note, noteRes)
	})

	t.Run("Test Delete Note", func(t *testing.T) {
		w := httptest.NewRecorder()
		tm, err := time.Parse("02.01.2006 15:04", "14.01.2024 11:03")
		assert.NoError(t, err)
		note := models.Note{
			ID:          1,
			Title:       "test",
			Description: "test",
			DateAdded:   tm,
			DateNotify:  tm,
		}

		mockStr.On("DeleteNote", ctx, note.ID).Return(nilError)

		req, err := http.NewRequestWithContext(ctx, "DELETE", "/notes/?id=1", nil)
		assert.NoError(t, err)

		serv.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
		assert.NoError(t, err)
	})

	t.Run("Test Update Note", func(t *testing.T) {
		w := httptest.NewRecorder()
		tm, err := time.Parse("02.01.2006 15:04", "14.01.2024 11:03")
		assert.NoError(t, err)
		note := models.Note{
			ID:          1,
			Title:       "test",
			Description: "test",
			DateAdded:   tm,
			DateNotify:  tm,
		}

		mockStr.On("UpdateNote", ctx, note).Return(nilError)

		b, err := json.Marshal(note)
		assert.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, "PATCH", "/notes/", bytes.NewReader(b))
		assert.NoError(t, err)

		serv.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
		assert.NoError(t, err)
	})

	mockStr.AssertExpectations(t)
}
