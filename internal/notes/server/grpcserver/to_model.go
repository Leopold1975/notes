package grpcserver

import (
	"notes/internal/notes/server/grpcserver/pb"
	"notes/internal/pkg/models"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToNote(n *pb.Note) models.Note {
	return models.Note{
		ID:          n.ID,
		Title:       n.Title,
		Description: n.Description,
		DateAdded:   n.DateAdded.AsTime(),
		DateNotify:  n.DateNotify.AsTime(),
	}
}

func ToNotes(notesPB []*pb.Note) []models.Note {
	notes := make([]models.Note, 0, len(notesPB))
	for _, v := range notesPB {
		notes = append(notes, ToNote(v))
	}
	return notes
}

func ToPBNotes(notes []models.Note) []*pb.Note {
	notesPB := make([]*pb.Note, 0, len(notes))

	for _, n := range notes {
		notesPB = append(notesPB, ToPBNote(n))
	}

	return notesPB
}

func ToPBNote(n models.Note) *pb.Note {
	return &pb.Note{
		ID:          n.ID,
		Title:       n.Title,
		Description: n.Description,
		DateAdded:   timestamppb.New(n.DateAdded),
		DateNotify:  timestamppb.New(n.DateNotify),
	}
}
