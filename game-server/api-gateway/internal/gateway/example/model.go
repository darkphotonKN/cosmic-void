package example

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/example"
)

type ExampleClient interface {
	CreateExample(ctx context.Context, req *pb.CreateExampleRequest) (*pb.Example, error)
	GetExample(ctx context.Context, req *pb.GetExampleRequest) (*pb.Example, error)
}
