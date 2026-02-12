package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	authpb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/darkphotonKN/cosmic-void-server/common/utils/cache"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/subscription"
	"github.com/stripe/stripe-go/v82/webhook"
)

type Service interface {
	CreateCustomer(ctx context.Context, userId uuid.UUID, email string) (*CreateCustomerResponse, error)
	SetupSubscription(ctx context.Context, req *SetupSubscriptionReq) (*SetupSubscriptionResp, error)
	Subscribe(ctx context.Context, req *SubscribeRequest) (*SubscribeResponse, error)
	GetUserSubscriptions(ctx context.Context, customerID string) (*UserSubscriptionsResponse, error)
	ProcessWebhookEvent(ctx context.Context, payload []byte, signature string) error
	CheckPermission(ctx context.Context, userID string) (bool, error)
}

type service struct {
	repo           Repository
	processor      PaymentProcessor
	publishCh      *amqp.Channel
	registry       discovery.Registry
	cache          cache.Cache
	webhookSecret  string
}

func NewService(repo Repository, processor PaymentProcessor, ch *amqp.Channel, registry discovery.Registry, cache cache.Cache, webhookSecret string) Service {
	return &service{
		repo:          repo,
		processor:     processor,
		publishCh:     ch,
		registry:      registry,
		cache:         cache,
		webhookSecret: webhookSecret,
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

	// Cache customer → user mapping in Redis
	cacheKey := fmt.Sprintf("customer:user_mapping:%s", customerID)
	if err := s.cache.Set(ctx, cacheKey, userId.String(), 24*time.Hour); err != nil {
		slog.Warn("Failed to cache customer:user_mapping", "error", err)
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

// ProcessWebhookEvent verifies the Stripe webhook signature, parses the event,
// and syncs subscription state from Stripe to the local DB.
func (s *service) ProcessWebhookEvent(ctx context.Context, payload []byte, signature string) error {
	// 1. Verify webhook signature
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return fmt.Errorf("webhook signature verification failed: %w", err)
	}

	slog.Info("Processing webhook event", "type", event.Type)

	// 2. Extract customer ID from event
	customerID, err := extractCustomerID(event)
	if err != nil {
		slog.Warn("Could not extract customer ID from event", "type", event.Type, "error", err)
		return nil // non-critical, return success to Stripe
	}

	if customerID == "" {
		slog.Info("No customer ID in event, skipping", "type", event.Type)
		return nil
	}

	// 3. Look up user ID via Redis cache or auth-service
	userID, err := s.getUserIDByCustomerID(ctx, customerID)
	if err != nil || userID == "" {
		slog.Warn("Could not find user for customer", "customerID", customerID, "error", err)
		return nil
	}

	// 4. Fetch ALL subscriptions from Stripe for this customer and sync to DB
	if err := s.syncSubscriptions(ctx, customerID, userID); err != nil {
		slog.Error("Failed to sync subscriptions", "error", err)
		return err
	}

	return nil
}

// extractCustomerID pulls the customer ID from different Stripe event types.
func extractCustomerID(event stripe.Event) (string, error) {
	switch event.Type {
	case "customer.subscription.created",
		"customer.subscription.updated",
		"customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return "", err
		}
		return sub.Customer.ID, nil

	case "invoice.payment_succeeded",
		"invoice.payment_failed":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			return "", err
		}
		return inv.Customer.ID, nil

	case "payment_intent.succeeded",
		"payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return "", err
		}
		return pi.Customer.ID, nil

	default:
		return "", nil
	}
}

// getUserIDByCustomerID checks Redis cache first, then falls back to auth-service.
func (s *service) getUserIDByCustomerID(ctx context.Context, customerID string) (string, error) {
	cacheKey := fmt.Sprintf("customer:user_mapping:%s", customerID)

	// Try Redis cache first
	userID, err := s.cache.Get(ctx, cacheKey)
	if err == nil && userID != "" {
		return userID, nil
	}

	// Cache miss — look up via Stripe customer metadata
	// The customer was created with metadata["user_id"]
	conn, err := discovery.ServiceConnection(ctx, "auth", s.registry)
	if err != nil {
		return "", fmt.Errorf("failed to connect to auth service: %w", err)
	}
	defer conn.Close()

	// We don't have a "get member by customer ID" RPC, so we'll look it up
	// from Stripe customer metadata directly
	cust, err := getStripeCustomer(customerID)
	if err != nil {
		return "", fmt.Errorf("failed to get stripe customer: %w", err)
	}

	userIDFromMeta, ok := cust.Metadata["user_id"]
	if !ok || userIDFromMeta == "" {
		return "", fmt.Errorf("no user_id in customer metadata for %s", customerID)
	}

	// Update cache for future lookups
	if cacheErr := s.cache.Set(ctx, cacheKey, userIDFromMeta, 24*time.Hour); cacheErr != nil {
		slog.Warn("Failed to cache customer:user_mapping", "error", cacheErr)
	}

	return userIDFromMeta, nil
}

// syncSubscriptions fetches all subscriptions from Stripe, upserts them into the DB,
// and updates the member's subscription status in auth-service.
func (s *service) syncSubscriptions(ctx context.Context, customerID, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user UUID: %w", err)
	}

	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
	}
	params.AddExpand("data.items.data.price")

	var latestSub *stripe.Subscription
	var latestCreated int64

	iter := subscription.List(params)
	for iter.Next() {
		sub := iter.Subscription()

		localSub := &Subscription{
			ID:                   uuid.New(),
			UserID:               uid,
			StripeCustomerID:     customerID,
			StripeSubscriptionID: sub.ID,
			Status:               string(sub.Status),
			CancelAtPeriodEnd:    sub.CancelAtPeriodEnd,
		}

		if sub.CanceledAt > 0 {
			t := time.Unix(sub.CanceledAt, 0)
			localSub.CanceledAt = &t
		}

		// Get price ID and period from first item
		if len(sub.Items.Data) > 0 {
			item := sub.Items.Data[0]
			localSub.StripePriceID = item.Price.ID
			localSub.CurrentPeriodStart = time.Unix(item.CurrentPeriodStart, 0)
			localSub.CurrentPeriodEnd = time.Unix(item.CurrentPeriodEnd, 0)
		}

		if err := s.repo.UpsertSubscription(ctx, localSub); err != nil {
			slog.Error("Failed to upsert subscription", "subID", sub.ID, "error", err)
			continue
		}

		// Track the latest subscription by creation time
		if sub.Created > latestCreated {
			latestCreated = sub.Created
			latestSub = sub
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("error listing subscriptions: %w", err)
	}

	// Update member's subscription status with the latest subscription
	if latestSub != nil {
		productID := ""
		if len(latestSub.Items.Data) > 0 && latestSub.Items.Data[0].Price.Product != nil {
			productID = latestSub.Items.Data[0].Price.Product.ID
		}

		if err := s.updateMemberSubscriptionStatus(ctx, userID, productID, string(latestSub.Status)); err != nil {
			slog.Error("Failed to update member subscription status", "error", err)
		}

		// Cache user permission in Redis for fast polling lookups
		hasPermission := latestSub.Status == stripe.SubscriptionStatusActive ||
			latestSub.Status == stripe.SubscriptionStatusTrialing
		cacheKey := fmt.Sprintf("user:permission:%s", userID)
		permValue := "false"
		if hasPermission {
			permValue = "true"
		}
		if err := s.cache.Set(ctx, cacheKey, permValue, 0); err != nil {
			slog.Warn("Failed to cache user permission", "error", err)
		}
	}

	slog.Info("Synced subscriptions from Stripe", "customerID", customerID, "userID", userID)
	return nil
}

// CheckPermission checks if a user has an active subscription.
// Redis first, then falls back to the local subscriptions DB.
func (s *service) CheckPermission(ctx context.Context, userID string) (bool, error) {
	cacheKey := fmt.Sprintf("user:permission:%s", userID)

	// Try Redis cache first
	val, err := s.cache.Get(ctx, cacheKey)
	if err == nil && val != "" {
		return val == "true", nil
	}

	// Cache miss — check local subscriptions DB
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user UUID: %w", err)
	}

	subs, err := s.repo.GetSubscriptionsByUserID(ctx, uid)
	if err != nil {
		return false, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	hasPermission := false
	for _, sub := range subs {
		if sub.Status == string(stripe.SubscriptionStatusActive) ||
			sub.Status == string(stripe.SubscriptionStatusTrialing) {
			hasPermission = true
			break
		}
	}

	// Cache the result
	permValue := "false"
	if hasPermission {
		permValue = "true"
	}
	if cacheErr := s.cache.Set(ctx, cacheKey, permValue, 0); cacheErr != nil {
		slog.Warn("Failed to cache user permission", "error", cacheErr)
	}

	return hasPermission, nil
}

// updateMemberSubscriptionStatus calls auth-service to update the member's subscription fields.
func (s *service) updateMemberSubscriptionStatus(ctx context.Context, memberID, productID, status string) error {
	conn, err := discovery.ServiceConnection(ctx, "auth", s.registry)
	if err != nil {
		return fmt.Errorf("failed to connect to auth service: %w", err)
	}
	defer conn.Close()

	client := authpb.NewAuthServiceClient(conn)
	_, err = client.UpdateSubscriptionStatus(ctx, &authpb.UpdateSubscriptionStatusRequest{
		MemberId:  memberID,
		ProductId: productID,
		Status:    status,
	})
	return err
}
