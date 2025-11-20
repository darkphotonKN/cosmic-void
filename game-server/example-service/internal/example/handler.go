package example

import (
	"golang.org/x/net/context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/example"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
	pb.UnimplementedExampleServiceServer
}

type Service interface {
	CreateExample(ctx context.Context, example *pb.CreateExampleRequest) (*pb.Example, error)
	GetExample(ctx context.Context, id uuid.UUID) (*pb.Example, error)
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateExample(ctx context.Context, req *pb.CreateExampleRequest) (*pb.Example, error) {
	result, err := h.service.CreateExample(ctx, req)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (h *Handler) GetExample(ctx context.Context, req *pb.GetExampleRequest) (*pb.Example, error) {
	id, err := uuid.Parse(req.Id)

	if err != nil {
		return nil, err
	}

	result, err := h.service.GetExample(ctx, id)

	if err != nil {
		return nil, err
	}

	return result, nil
}
