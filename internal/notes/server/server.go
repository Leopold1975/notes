package server

import (
	"context"
	"notes/internal/notes/app"
	"notes/internal/pkg/models"
)

type App interface {
	app.Storage
	RefreshNote(context.Context, models.Note) error
}
