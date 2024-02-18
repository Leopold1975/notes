package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"notes/internal/notes/storage"
	"notes/internal/pkg/config"
	"notes/internal/pkg/models"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // used for driver
	"github.com/pressly/goose/v3"
)

// TODO: DB requests should create their own context with timeout, which is set dut to config.
type Storage struct {
	db *pgx.Conn
}

func New(ctx context.Context, cfg config.Config) (*Storage, error) {
	dbURL := "postgres://" + cfg.DB.Username + ":" + cfg.DB.Password + "@" +
		cfg.DB.Host + cfg.DB.Port + "/" + cfg.DB.DB
	db, err := connect(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	if err = applyMigrations(dbURL, cfg); err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func connect(ctx context.Context, dbURL string) (*pgx.Conn, error) {
	db, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	n := 1e9
loop:
	for {
		select {
		case <-ctx.Done():
			return nil, storage.ErrContextCancelled
		default:
			err = db.Ping(ctx)
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

func applyMigrations(dbURL string, cfg config.Config) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	defer db.Close()
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
	query := `INSERT INTO notes(title, description, date_added, date_notify, delay) 
	VALUES ($1, $2, $3, $4, $5)`

	_, err := s.db.Exec(
		ctx, query,
		note.Title,
		note.Description,
		note.DateAdded,  // Format("2006-01-02T15:04:05-07:00")
		note.DateNotify, // Format("2006-01-02T15:04:05-07:00")
		note.Delay,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetNotes(ctx context.Context, interval time.Duration) ([]models.Note, error) {
	querySq := squirrel.Select("id", "title", "description", "date_added", "date_notify", "delay").From("notes")
	if interval > 0 {
		querySq = querySq.Where(squirrel.And{
			squirrel.Expr(fmt.Sprintf("date_notify < NOW() +  '%s'", interval.String())),
			squirrel.Expr("date_notify >= NOW()"),
		})
	}
	query, _, err := querySq.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, query)
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
		err := rows.Scan(&n.ID, &n.Title, &n.Description, &n.DateAdded, &n.DateNotify, &n.Delay)
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
	query := "SELECT id, title, description, date_added, date_notify, delay FROM notes WHERE id = $1"

	n := models.Note{}
	if err := s.db.QueryRow(ctx, query, id).Scan(
		&n.ID, &n.Title, &n.Description, &n.DateAdded, &n.DateNotify, &n.Delay); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Note{}, storage.ErrNotFound
		}
		return models.Note{}, err
	}

	return n, nil
}

func (s *Storage) DeleteNote(ctx context.Context, id uint64) error {
	if id == 0 {
		return storage.ErrFieldUnspecified
	}
	query := `DELETE FROM notes WHERE id = $1`
	if _, err := s.db.Exec(ctx, query, id); err != nil {
		return err
	}

	return nil
}

func (s *Storage) UpdateNote(ctx context.Context, note models.Note) error {
	if note.ID == 0 {
		return storage.ErrFieldUnspecified
	}
	if _, err := s.GetNote(ctx, note.ID); errors.Is(err, storage.ErrNotFound) {
		return storage.ErrNotFound
	}
	qr := squirrel.Update("notes")

	fields := map[string]interface{}{
		"title":       note.Title,
		"description": note.Description,
		"date_notify": note.DateNotify,
		"delay":       note.Delay,
	}
	for field, value := range fields {
		// // Not really good via reflection.
		// if value != nil &&
		// 	!reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface()) {
		// 	qr = qr.Set(field, value)
		// }
		switch v := value.(type) {
		case string:
			if v == "" {
				continue
			}
		case time.Time:
			if v.IsZero() {
				continue
			}
		case time.Duration:
			if v == 0 {
				continue
			}
		}
		qr = qr.Set(field, value)
	}

	qr = qr.Where(squirrel.Eq{"id": note.ID}).PlaceholderFormat(squirrel.Dollar)
	q, args, err := qr.ToSql()
	if err != nil {
		log.Println(len(args))
		if len(args) == 0 {
			return fmt.Errorf("%w, %w", storage.ErrNotEnoughArguments, err)
		}
		return err
	}

	if _, err := s.db.Exec(ctx, q, args...); err != nil {
		return err
	}
	return nil
}
