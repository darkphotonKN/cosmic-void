package grpcauth

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
)

type AuthClient interface {
	GetMember(ctx context.Context, req *pb.GetMemberRequest) (*pb.Member, error)
	ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error)
}
