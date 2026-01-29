package payment

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/payment"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
	pb.UnimplementedPaymentServiceServer
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateCustomer(ctx context.Context, req *pb.CreateCustomerRequest) (*pb.CreateCustomerResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	resp, err := h.service.CreateCustomer(ctx, userId, req.Email)
	if err != nil {
		return nil, err
	}

	return &pb.CreateCustomerResponse{
		CustomerId: resp.CustomerID,
	}, nil
}

func (h *Handler) SetupSubscription(ctx context.Context, req *pb.SetupSubscriptionRequest) (*pb.SetupSubscriptionResponse, error) {
	resp, err := h.service.SetupSubscription(ctx, &SetupSubscriptionReq{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
	})
	if err != nil {
		return nil, err
	}

	return &pb.SetupSubscriptionResponse{
		ProductId: resp.ProductID,
		PriceId:   resp.PriceID,
	}, nil
}

func (h *Handler) Subscribe(ctx context.Context, req *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
	resp, err := h.service.Subscribe(ctx, &SubscribeRequest{
		ProductID:  req.ProductId,
		CustomerID: req.CustomerId,
	})
	if err != nil {
		return nil, err
	}

	return &pb.SubscribeResponse{
		SubscriptionId: resp.SubscriptionID,
		ClientSecret:   resp.ClientSecret,
		Status:         resp.Status,
	}, nil
}

func (h *Handler) GetUserSubscriptions(ctx context.Context, req *pb.GetUserSubscriptionsRequest) (*pb.GetUserSubscriptionsResponse, error) {
	resp, err := h.service.GetUserSubscriptions(ctx, req.CustomerId)
	if err != nil {
		return nil, err
	}

	var subscriptions []*pb.UserSubscriptionInfo
	for _, sub := range resp.Subscriptions {
		subscriptions = append(subscriptions, &pb.UserSubscriptionInfo{
			SubscriptionId:   sub.SubscriptionID,
			ProductId:        sub.ProductID,
			PriceId:          sub.PriceID,
			Status:           sub.Status,
			CurrentPeriodEnd: sub.CurrentPeriodEnd,
		})
	}

	return &pb.GetUserSubscriptionsResponse{
		Subscriptions: subscriptions,
	}, nil
}
