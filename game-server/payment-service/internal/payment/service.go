package payment

import (
	"context"

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
}

func NewService(repo Repository, processor PaymentProcessor, ch *amqp.Channel) Service {
	return &service{
		repo:      repo,
		processor: processor,
		publishCh: ch,
	}
}

func (s *service) CreateCustomer(ctx context.Context, userId uuid.UUID, email string) (*CreateCustomerResponse, error) {
	customerID, err := s.processor.CreateCustomer(ctx, userId, email)
	if err != nil {
		return nil, err
	}

	// TODO: Save customer ID to database via repo

	return &CreateCustomerResponse{
		CustomerID: customerID,
	}, nil
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
