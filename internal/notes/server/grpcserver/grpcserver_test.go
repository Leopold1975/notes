package grpcserver_test

import (
	"context"
	"net"
	"testing"
	"time"

	"notes/internal/notes/app"
	"notes/internal/notes/server/grpcserver"
	"notes/internal/notes/server/grpcserver/pb"
	"notes/internal/notes/storage"
	"notes/internal/pkg/config"
	"notes/internal/pkg/logger"
	"notes/internal/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var nilError error

func TestBasic(t *testing.T) {
	mockStr := new(storage.MockStorage)
	logg, err := logger.New(logger.EnvLocal)
	require.NoError(t, err)

	mockApp := app.NewApp(mockStr)

	server := grpcserver.New(mockApp, logg, config.GRPCServer{})
	s := grpc.NewServer()

	pb.RegisterNotesServer(s, server)

	lis := bufconn.Listen(1024 * 1024)
	go func() {
		err := s.Serve(lis)
		require.NoError(t, err)
	}()
	defer s.Stop()
	defer lis.Close()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	defer conn.Close()

	client := pb.NewNotesClient(conn)

	t.Parallel()

	ctx := context.Background()

	t.Run("Test Get Notes", func(t *testing.T) {
		exp := []models.Note{}
		var td time.Duration
		mockStr.On("GetNotes", ctx, td).Return(exp, nilError)

		res, err := client.GetNotes(ctx, &pb.GetNotesRequest{})

		assert.NoError(t, err)
		assert.Len(t, res.GetNotes(), 0)
	})

	t.Run("Test Create Note", func(t *testing.T) {
		tm, err := time.Parse("02.01.2006 15:04", "14.01.2024 11:03")
		assert.NoError(t, err)
		note := models.Note{
			Title:       "test",
			Description: "test",
			DateAdded:   tm,
			DateNotify:  tm,
		}

		mockStr.On("CreateNote", ctx, note).Return(nilError)

		_, err = client.CreateNote(ctx, &pb.CreateNoteRequest{
			Note: &pb.Note{
				Title:       "test",
				Description: "test",
				DateAdded:   timestamppb.New(tm),
				DateNotify:  timestamppb.New(tm),
			},
		})
		assert.NoError(t, err)
	})

	t.Run("Test Get Note", func(t *testing.T) {
		tm, err := time.Parse("02.01.2006 15:04", "14.01.2024 11:03")
		assert.NoError(t, err)
		note := models.Note{
			ID:          1,
			Title:       "test",
			Description: "test",
			DateAdded:   tm,
			DateNotify:  tm,
		}

		ctx := context.Background()
		mockStr.On("GetNote", ctx, note.ID).Return(note, nilError)

		res, err := client.GetNote(ctx, &pb.GetNoteRequest{
			ID: 1,
		})
		assert.NoError(t, err)
		assert.Equal(t, note, grpcserver.ToNote(res.Note))
	})

	t.Run("Test Delete Note", func(t *testing.T) {
		tm, err := time.Parse("02.01.2006 15:04", "14.01.2024 11:03")
		assert.NoError(t, err)
		note := models.Note{
			ID:          1,
			Title:       "test",
			Description: "test",
			DateAdded:   tm,
			DateNotify:  tm,
		}

		ctx := context.Background()
		mockStr.On("DeleteNote", ctx, note.ID).Return(nilError)

		_, err = client.DeleteNote(ctx, &pb.DeleteNoteRequest{
			ID: note.ID,
		})

		assert.NoError(t, err)
	})

	t.Run("Test Update Note", func(t *testing.T) {
		tm, err := time.Parse("02.01.2006 15:04", "14.01.2024 11:03")
		assert.NoError(t, err)
		note := models.Note{
			ID:          1,
			Title:       "test",
			Description: "test",
			DateAdded:   tm,
			DateNotify:  tm,
		}

		ctx := context.Background()
		mockStr.On("UpdateNote", ctx, note).Return(nilError)

		_, err = client.UpdateNote(ctx, &pb.UpdateNoteRequest{
			Note: grpcserver.ToPBNote(note),
		})
		assert.NoError(t, err)
	})

	mockStr.AssertExpectations(t)
}
