package payment

import (
	"context"
	"fmt"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/payment"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
)

const (
	serviceName = "payments"
)

type Client struct {
	registry discovery.Registry
}

func NewClient(registry discovery.Registry) PaymentClient {
	return &Client{
		registry: registry,
	}
}

func (c *Client) CreateCustomer(ctx context.Context, req *pb.CreateCustomerRequest) (*pb.CreateCustomerResponse, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}
	defer conn.Close()

	client := pb.NewPaymentServiceClient(conn)
	return client.CreateCustomer(ctx, req)
}

func (c *Client) SetupSubscription(ctx context.Context, req *pb.SetupSubscriptionRequest) (*pb.SetupSubscriptionResponse, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}
	defer conn.Close()

	client := pb.NewPaymentServiceClient(conn)
	return client.SetupSubscription(ctx, req)
}

func (c *Client) Subscribe(ctx context.Context, req *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}
	defer conn.Close()

	client := pb.NewPaymentServiceClient(conn)
	return client.Subscribe(ctx, req)
}

func (c *Client) GetUserSubscriptions(ctx context.Context, req *pb.GetUserSubscriptionsRequest) (*pb.GetUserSubscriptionsResponse, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}
	defer conn.Close()

	client := pb.NewPaymentServiceClient(conn)
	return client.GetUserSubscriptions(ctx, req)
}
