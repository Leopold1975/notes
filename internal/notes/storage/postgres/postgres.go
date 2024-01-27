package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // used for driver
	"github.com/pressly/goose/v3"

	"notes/internal/notes/storage"
	"notes/internal/pkg/config"
	"notes/internal/pkg/models"
)

// TODO: DB requests should create their own context with timeout, which is set dut to config.
type Storage struct {
	db *sql.DB
}

func New(ctx context.Context, cfg config.Config) (*Storage, error) {
	// log = log.With("localtion", "storage/postgres/postgres.go")

	db, err := connect(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err = applyMigrations(db, cfg); err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func connect(ctx context.Context, cfg config.Config) (*sql.DB, error) {
	dbURL := "postgres://" + cfg.DB.Username + ":" + cfg.DB.Password + "@" +
		cfg.DB.Host + cfg.DB.Port + "/" + cfg.DB.DB

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, err
	}

	n := 1e6
loop:
	for {
		select {
		case <-ctx.Done():
			return nil, storage.ErrContextCancelled
		default:
			err = db.Ping()
			if err == nil {
				break loop
			}

			target := new(pgconn.PgError)
			if errors.As(err, &target) {
				if target.Code == "3D000" {
					return nil, err
				}
			}

			log.Printf("connecting to db... error %s\n", err.Error())
			time.Sleep(time.Duration(n))
			n += 3e9
		}
	}
	return db, nil
}

func applyMigrations(db *sql.DB, cfg config.Config) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if cfg.DB.Reload {
		if err := goose.DownTo(db, "./migrations", 0); err != nil {
			return err
		}
	}

	// version 1 is a default version that just creates table if it doesn't exist.
	if err := goose.UpTo(db, "./migrations", cfg.DB.Version); err != nil {
		return err
	}
	return nil
}

func (s *Storage) CreateNote(ctx context.Context, note models.Note) error {
	// TODO: can return id so that we can add the note to cache i.e. Redis to have
	// access to it without requesting db, like this: redisDB.Add(Key: id, Value: note).
	query := `INSERT INTO notes(title, description, date_added, date_notify) 
	VALUES ($1, $2, $3, $4)`
	row := s.db.QueryRowContext(
		ctx, query, note.Title, note.Description,
		time.Now().Format("2006-01-02 15:04:05-0700"), note.DateNotify,
	)
	if row.Err() != nil {
		return row.Err()
	}

	return nil
}

func (s *Storage) GetNotes(ctx context.Context) ([]models.Note, error) {
	query := `SELECT id, title, description, date_added, date_notify FROM notes`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}

	defer rows.Close()
	notes := make([]models.Note, 0, 32)
	for rows.Next() {
		n := models.Note{}
		err := rows.Scan(&n.ID, &n.Title, &n.Description, &n.DateAdded, &n.DateNotify)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

func (s *Storage) GetNote(ctx context.Context, id uint64) (models.Note, error) {
	if id == 0 {
		return models.Note{}, storage.ErrFieldUnspecified
	}
	query := "SELECT id, title, description, date_added, date_notify FROM notes WHERE id = $1"
	row := s.db.QueryRowContext(ctx, query, id)

	if row.Err() != nil {
		return models.Note{}, row.Err()
	}

	n := models.Note{}
	if err := row.Scan(&n.ID, &n.Title, &n.Description, &n.DateAdded, &n.DateNotify); err != nil {
		return models.Note{}, err
	}
	return n, nil
}

func (s *Storage) DeleteNote(ctx context.Context, id uint64) error {
	if id == 0 {
		return storage.ErrFieldUnspecified
	}
	query := `DELETE FROM notes WHERE id = $1`
	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}

	return nil
}

func (s *Storage) UpdateNote(ctx context.Context, note models.Note) error {
	if note.ID == 0 {
		return storage.ErrFieldUnspecified
	}
	query := `UPDATE notes SET title = $1, description = $2, date_notify = $3 WHERE id = $4`
	rows, err := s.db.QueryContext(ctx, query, note.Title, note.Description, note.DateNotify, note.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}
	return nil
}
