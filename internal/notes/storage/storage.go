package storage

import (
	"context"
	"errors"

	"notes/internal/pkg/models"
)

var (
	ErrContextCancelled  = errors.New("context cancelled connection")
	ErrFieldUnspecified  = errors.New("required fields are unspecified")
	ErrDatabaseNotExists = errors.New("database don't exist")
)

type NoteCreater interface {
	CreateNote(context.Context, models.Note) error
}

type NotesGetter interface {
	GetNotes(context.Context) ([]models.Note, error)
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
