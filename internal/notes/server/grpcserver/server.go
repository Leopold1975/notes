package grpcserver

import (
	"context"
	"errors"
	"net"
	"notes/internal/notes/server"
	"notes/internal/notes/server/grpcserver/interceptor"
	"notes/internal/notes/server/grpcserver/pb"
	"notes/internal/notes/storage"
	"notes/internal/pkg/config"
	"notes/internal/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	a      server.App
	cfg    config.GRPCServer
	server *grpc.Server
	pb.NotesServer
}

func New(a server.App, logg logger.Logger, cfg config.GRPCServer) *Server {
	return &Server{
		a:   a,
		cfg: cfg,
		server: grpc.NewServer(
			grpc.UnaryInterceptor(
				grpc.UnaryServerInterceptor(
					interceptor.LoggingInterceptor(logg)),
			),
		),
	}
}

func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.cfg.Host+s.cfg.Port)
	if err != nil {
		return err
	}

	defer lis.Close()
	pb.RegisterNotesServer(s.server, s)

	select {
	case <-ctx.Done():
		return nil
	default:
		if err := s.server.Serve(lis); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) Shutdown(_ context.Context) error {
	s.server.GracefulStop()
	return nil
}

func (s *Server) GetNotes(ctx context.Context, r *pb.GetNotesRequest) (*pb.GetNotesResponse, error) {
	notes, err := s.a.GetNotes(ctx, r.TimeInterval.AsDuration())
	if err != nil {
		return &pb.GetNotesResponse{}, status.Error(codes.Internal, err.Error())
	}

	pbNotes := ToPBNotes(notes)
	return &pb.GetNotesResponse{Notes: pbNotes}, nil
}

func (s *Server) GetNote(ctx context.Context, req *pb.GetNoteRequest) (*pb.GetNoteResponse, error) {
	note, err := s.a.GetNote(ctx, req.ID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return &pb.GetNoteResponse{}, status.Error(codes.NotFound, err.Error())
		}
		return &pb.GetNoteResponse{}, status.Error(codes.Internal, err.Error())
	}

	notePB := ToPBNote(note)
	return &pb.GetNoteResponse{Note: notePB}, nil
}

func (s *Server) CreateNote(ctx context.Context, req *pb.CreateNoteRequest) (*pb.CreateNoteResponse, error) {
	note := ToNote(req.Note)
	if err := s.a.CreateNote(ctx, note); err != nil {
		return &pb.CreateNoteResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &pb.CreateNoteResponse{}, nil
}

func (s *Server) DeleteNote(ctx context.Context, req *pb.DeleteNoteRequest) (*pb.DeleteNoteResponse, error) {
	if err := s.a.DeleteNote(ctx, req.ID); err != nil {
		return &pb.DeleteNoteResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &pb.DeleteNoteResponse{}, nil
}

func (s *Server) UpdateNote(ctx context.Context, req *pb.UpdateNoteRequest) (*pb.UpdateNoteResponse, error) {
	note := ToNote(req.Note)
	md, _ := metadata.FromIncomingContext(ctx)

	r := md.Get("refreshed")
	if len(r) > 0 {
		if r[0] != "true" {
			return &pb.UpdateNoteResponse{}, status.Error(codes.InvalidArgument, "invalid metadata")
		}
		if err := s.a.RefreshNote(ctx, note); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return &pb.UpdateNoteResponse{}, status.Error(codes.NotFound, err.Error())
			}
			return &pb.UpdateNoteResponse{}, status.Error(codes.Internal, err.Error())
		}
	}

	if err := s.a.UpdateNote(ctx, note); err != nil {
		return &pb.UpdateNoteResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdateNoteResponse{}, nil
}
