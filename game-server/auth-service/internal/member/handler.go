package member

import (
	"context"
	"fmt"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
)

type Handler struct {
	pb.UnimplementedAuthServiceServer
	service Service
}

type Service interface {
	CreateMember(ctx context.Context, req *pb.CreateMemberRequest) (*pb.Member, error)
	GetMember(ctx context.Context, req *pb.GetMemberRequest) (*pb.Member, error)
	LoginMember(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error)
	UpdateMemberInfo(ctx context.Context, req *pb.UpdateMemberInfoRequest) (*pb.Member, error)
	UpdateMemberPassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error)
	ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error)
	CreateDefaultMembers(members []CreateDefaultMember) error
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (s *Handler) LoginMember(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return s.service.LoginMember(ctx, req)
}

func (s *Handler) GetMember(ctx context.Context, req *pb.GetMemberRequest) (*pb.Member, error) {
	return s.service.GetMember(ctx, req)
}

func (s *Handler) CreateMember(ctx context.Context, req *pb.CreateMemberRequest) (*pb.Member, error) {
	fmt.Printf("Creating member through auth-service, request: %+v\n", req)
	return s.service.CreateMember(ctx, req)
}

func (s *Handler) UpdateMemberInfo(ctx context.Context, req *pb.UpdateMemberInfoRequest) (*pb.Member, error) {
	return s.service.UpdateMemberInfo(ctx, req)
}

func (s *Handler) UpdateMemberPassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	return s.service.UpdateMemberPassword(ctx, req)
}

func (s *Handler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	return s.service.ValidateToken(ctx, req)
}
