package publisher

import (
	"context"
	"fmt"

	"notes/internal/pkg/logger"
	"notes/internal/pkg/models"
)

type SenderShutdowner interface {
	Send(context.Context, models.Message) error
	Shutdown() error
}

type Fetcher interface {
	Fetch(context.Context) ([]models.Note, error)
}

type Publisher struct {
	s SenderShutdowner
	f Fetcher
	l logger.Logger
}

func New(s SenderShutdowner, f Fetcher, logg logger.Logger) (Publisher, error) {
	return Publisher{
		s: s,
		f: f,
		l: logg,
	}, nil
}

func (ks *Publisher) Publish(ctx context.Context) error {
	notes, err := ks.f.Fetch(ctx)
	if err != nil {
		return err
	}

	for _, n := range notes {
		m, err := models.NoteToMessage(n)
		if err != nil {
			ks.l.Error(fmt.Errorf("can not convert note to message error: %w", err).Error())
			continue
		}

		err = ks.s.Send(ctx, m)
		if err != nil {
			ks.l.Error(fmt.Errorf("can not send message to queque error: %w", err).Error())
			continue
		}
	}
	return nil
}

func (ks *Publisher) Shutdown(ctx context.Context) error {
	ok := make(chan error)
	defer close(ok)
	go func() {
		ok <- ks.s.Shutdown()
		close(ok)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("shutdown context timeout exceeded")
	case err := <-ok:
		if err != nil {
			return err
		}
		return nil
	}
}
