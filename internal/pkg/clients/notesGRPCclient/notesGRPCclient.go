package notesgrpcclient

import (
	"context"
	"notes/internal/notes/server/grpcserver"
	"notes/internal/notes/server/grpcserver/pb"
	"notes/internal/pkg/config"
	"notes/internal/pkg/models"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Client struct {
	cl   pb.NotesClient
	conn *grpc.ClientConn
}

func New(cfg config.GRPCServer) (*Client, error) {
	conn, err := grpc.Dial(cfg.Host+cfg.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{pb.NewNotesClient(conn), conn}, err
}

func (c *Client) Fetch(ctx context.Context) ([]models.Note, error) {
	notes, err := c.GetNotes(ctx, time.Minute*5)
	if err != nil {
		return nil, err
	}
	return notes, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) UpdateNote(ctx context.Context, note models.Note) error {
	v := ctx.Value("refreshed")
	vv, ok := v.(bool)
	if ok && vv {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
			"refreshed": "true",
		}))
	}

	res, err := c.cl.UpdateNote(ctx, &pb.UpdateNoteRequest{
		Note: grpcserver.ToPBNote(note),
	})
	if err != nil {
		return err
	}
	_ = res
	return nil
}

func (c *Client) CreateNote(ctx context.Context, note models.Note) error {
	n := grpcserver.ToPBNote(note)
	_, err := c.cl.CreateNote(ctx, &pb.CreateNoteRequest{
		Note: n,
	})
	return err
}

func (c *Client) GetNotes(ctx context.Context, interval time.Duration) ([]models.Note, error) {
	res, err := c.cl.GetNotes(ctx, &pb.GetNotesRequest{
		TimeInterval: durationpb.New(interval),
	})
	if err != nil {
		return nil, err
	}
	return grpcserver.ToNotes(res.Notes), nil
}

func (c *Client) GetNote(ctx context.Context, id uint64) (models.Note, error) {
	res, err := c.cl.GetNote(ctx, &pb.GetNoteRequest{
		ID: id,
	})
	if err != nil {
		return models.Note{}, err
	}
	return grpcserver.ToNote(res.Note), nil
}
