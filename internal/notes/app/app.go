package app

import (
	"context"
	"notes/internal/pkg/models"
	"time"
)

type NoteCreater interface {
	CreateNote(context.Context, models.Note) error
}

type NotesGetter interface {
	GetNotes(context.Context, time.Duration) ([]models.Note, error)
}

type NoteDeleter interface {
	DeleteNote(context.Context, uint64) error
}

type NoteUpdater interface {
	UpdateNote(context.Context, models.Note) error
}

type NoteGetter interface {
	GetNote(context.Context, uint64) (models.Note, error)
}

type Storage interface {
	NoteCreater
	NotesGetter
	NoteGetter
	NoteDeleter
	NoteUpdater
}

type NotesApp struct {
	str Storage
	// cache
}

func NewApp(str Storage) *NotesApp {
	return &NotesApp{
		str: str,
	}
}

func (a *NotesApp) CreateNote(ctx context.Context, note models.Note) error {
	note.Delay = time.Minute * 20
	note.DateAdded = time.Now()
	note.DateNotify = note.DateAdded.Add(note.Delay)

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

func (a *NotesApp) RefreshNote(ctx context.Context, note models.Note) error {
	note, err := a.GetNote(ctx, note.ID)
	if err != nil {
		return err
	}
	note.DateNotify = note.DateNotify.Add(note.Delay)
	note.Delay *= 10

	if note.Delay > time.Hour*24*365 {
		return a.DeleteNote(ctx, note.ID)
	}
	return a.UpdateNote(ctx, note)
}
