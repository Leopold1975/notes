package storage

import (
	"context"

	"notes/internal/pkg/models"

	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (s *MockStorage) CreateNote(_ context.Context, n models.Note) error {
	ctx := context.Background()
	args := s.Called(ctx, n)

	return args.Error(0)
}

func (s *MockStorage) GetNotes(_ context.Context) ([]models.Note, error) {
	ctx := context.Background()

	args := s.Called(ctx)

	return args.Get(0).([]models.Note), args.Error(1)
}

func (s *MockStorage) GetNote(_ context.Context, id uint64) (models.Note, error) {
	ctx := context.Background()

	args := s.Called(ctx, id)

	return args.Get(0).(models.Note), args.Error(1)
}

func (s *MockStorage) DeleteNote(_ context.Context, id uint64) error {
	ctx := context.Background()

	args := s.Called(ctx, id)

	return args.Error(0)
}

func (s *MockStorage) UpdateNote(_ context.Context, n models.Note) error {
	ctx := context.Background()

	args := s.Called(ctx, n)

	return args.Error(0)
}
