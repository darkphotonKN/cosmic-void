package payment

import (
	"context"
	"fmt"
	"log/slog"

	authpb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Service interface {
	CreateCustomer(ctx context.Context, userId uuid.UUID, email string) (*CreateCustomerResponse, error)
	SetupSubscription(ctx context.Context, req *SetupSubscriptionReq) (*SetupSubscriptionResp, error)
	Subscribe(ctx context.Context, req *SubscribeRequest) (*SubscribeResponse, error)
	GetUserSubscriptions(ctx context.Context, customerID string) (*UserSubscriptionsResponse, error)
}

type service struct {
	repo      Repository
	processor PaymentProcessor
	publishCh *amqp.Channel
	registry  discovery.Registry
}

func NewService(repo Repository, processor PaymentProcessor, ch *amqp.Channel, registry discovery.Registry) Service {
	return &service{
		repo:      repo,
		processor: processor,
		publishCh: ch,
		registry:  registry,
	}
}

func (s *service) CreateCustomer(ctx context.Context, userId uuid.UUID, email string) (*CreateCustomerResponse, error) {
	// Check if the member already has a Stripe customer ID
	existingID, err := s.getStripeCustomerID(ctx, userId.String())
	if err != nil {
		slog.Warn("Failed to check existing stripe customer ID", "error", err)
	}
	if existingID != "" {
		return &CreateCustomerResponse{CustomerID: existingID}, nil
	}

	// Create new Stripe customer
	customerID, err := s.processor.CreateCustomer(ctx, userId, email)
	if err != nil {
		return nil, err
	}

	// Save customer ID to auth-service via gRPC
	if err := s.saveStripeCustomerID(ctx, userId.String(), customerID); err != nil {
		slog.Error("Failed to save stripe customer ID to auth-service", "error", err)
		return nil, fmt.Errorf("failed to save customer ID: %w", err)
	}

	return &CreateCustomerResponse{CustomerID: customerID}, nil
}

func (s *service) getStripeCustomerID(ctx context.Context, memberID string) (string, error) {
	conn, err := discovery.ServiceConnection(ctx, "auth", s.registry)
	if err != nil {
		return "", fmt.Errorf("failed to connect to auth service: %w", err)
	}
	defer conn.Close()

	client := authpb.NewAuthServiceClient(conn)
	resp, err := client.GetStripeCustomerID(ctx, &authpb.GetStripeCustomerIDRequest{
		MemberId: memberID,
	})
	if err != nil {
		return "", err
	}
	return resp.StripeCustomerId, nil
}

func (s *service) saveStripeCustomerID(ctx context.Context, memberID, customerID string) error {
	conn, err := discovery.ServiceConnection(ctx, "auth", s.registry)
	if err != nil {
		return fmt.Errorf("failed to connect to auth service: %w", err)
	}
	defer conn.Close()

	client := authpb.NewAuthServiceClient(conn)
	_, err = client.SetStripeCustomerID(ctx, &authpb.SetStripeCustomerIDRequest{
		MemberId:         memberID,
		StripeCustomerId: customerID,
	})
	return err
}

func (s *service) SetupSubscription(ctx context.Context, req *SetupSubscriptionReq) (*SetupSubscriptionResp, error) {
	return s.processor.SetupSubscription(ctx, req)
}

func (s *service) Subscribe(ctx context.Context, req *SubscribeRequest) (*SubscribeResponse, error) {
	resp, err := s.processor.SubscribeToProduct(ctx, req)
	if err != nil {
		return nil, err
	}

	// TODO: Save subscription to database via repo

	return resp, nil
}

func (s *service) GetUserSubscriptions(ctx context.Context, customerID string) (*UserSubscriptionsResponse, error) {
	return s.processor.GetUserSubscriptions(ctx, customerID)
}
