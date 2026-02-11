package payment

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/payment"
)

type PaymentClient interface {
	CreateCustomer(ctx context.Context, req *pb.CreateCustomerRequest) (*pb.CreateCustomerResponse, error)
	SetupSubscription(ctx context.Context, req *pb.SetupSubscriptionRequest) (*pb.SetupSubscriptionResponse, error)
	Subscribe(ctx context.Context, req *pb.SubscribeRequest) (*pb.SubscribeResponse, error)
	GetUserSubscriptions(ctx context.Context, req *pb.GetUserSubscriptionsRequest) (*pb.GetUserSubscriptionsResponse, error)
}
