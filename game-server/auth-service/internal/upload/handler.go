package upload

import (
	"context"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
)

type Handler struct {
	pb.UnimplementedUploadServiceServer
	service Service
}

type Service interface {
	RequestAvatarUpload(ctx context.Context, req *pb.RequestAvatarUploadRequest) (*pb.RequestAvatarUploadResponse, error)
	ConfirmAvatarUpload(ctx context.Context, req *pb.ConfirmAvatarUploadRequest) (*pb.ConfirmAvatarUploadResponse, error)
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (s *Handler) RequestAvatarUpload(ctx context.Context, req *pb.RequestAvatarUploadRequest) (*pb.RequestAvatarUploadResponse, error) {
	slog.Info("Requesting avatar upload", "member_id", req.MemberId, "filename", req.Filename)
	return s.service.RequestAvatarUpload(ctx, req)
}

func (s *Handler) ConfirmAvatarUpload(ctx context.Context, req *pb.ConfirmAvatarUploadRequest) (*pb.ConfirmAvatarUploadResponse, error) {
	slog.Info("Confirming avatar upload", "member_id", req.MemberId, "upload_id", req.UploadId)
	return s.service.ConfirmAvatarUpload(ctx, req)
}
