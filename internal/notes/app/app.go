package app

import (
	"context"
	"time"

	"notes/internal/notes/storage"
	"notes/internal/pkg/models"
)

type App interface {
	storage.Storage
}

type NotesApp struct {
	str storage.Storage
	// cache
}

func NewApp(str storage.Storage) App {
	return &NotesApp{
		str: str,
	}
}

func (a *NotesApp) CreateNote(ctx context.Context, note models.Note) error {
	return a.str.CreateNote(ctx, note)
}

func (a *NotesApp) GetNotes(ctx context.Context, interval time.Duration) ([]models.Note, error) {
	return a.str.GetNotes(ctx, interval)
}

func (a *NotesApp) GetNote(ctx context.Context, id uint64) (models.Note, error) {
	return a.str.GetNote(ctx, id)
}

func (a *NotesApp) DeleteNote(ctx context.Context, id uint64) error {
	return a.str.DeleteNote(ctx, id)
}

func (a *NotesApp) UpdateNote(ctx context.Context, note models.Note) error {
	return a.str.UpdateNote(ctx, note)
}
