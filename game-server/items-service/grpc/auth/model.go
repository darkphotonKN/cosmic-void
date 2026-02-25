package auth

import (
	"context"
	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
)

type AuthClient interface {
	GetMember(ctx context.Context, req *pb.GetMemberRequest) (*pb.Member, error)
}
