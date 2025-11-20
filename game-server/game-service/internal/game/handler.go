package game

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/game"
)

type Handler struct {
	pb.UnimplementedGameServiceServer
	service *service
}

func NewHandler(service *service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.Room, error) {
	return h.service.CreateRoom(ctx, req)
}

func (h *Handler) GetRoom(ctx context.Context, req *pb.GetRoomRequest) (*pb.Room, error) {
	return h.service.GetRoom(ctx, req)
}

func (h *Handler) ListRooms(ctx context.Context, req *pb.ListRoomsRequest) (*pb.ListRoomsResponse, error) {
	return h.service.ListRooms(ctx, req)
}

func (h *Handler) JoinRoom(ctx context.Context, req *pb.JoinRoomRequest) (*pb.JoinRoomResponse, error) {
	return h.service.JoinRoom(ctx, req)
}

func (h *Handler) LeaveRoom(ctx context.Context, req *pb.LeaveRoomRequest) (*pb.LeaveRoomResponse, error) {
	return h.service.LeaveRoom(ctx, req)
}

func (h *Handler) StartGame(ctx context.Context, req *pb.StartGameRequest) (*pb.StartGameResponse, error) {
	return h.service.StartGame(ctx, req)
}

func (h *Handler) EndGame(ctx context.Context, req *pb.EndGameRequest) (*pb.EndGameResponse, error) {
	return h.service.EndGame(ctx, req)
}