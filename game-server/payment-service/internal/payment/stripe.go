package payment

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/price"
	"github.com/stripe/stripe-go/v82/product"
	"github.com/stripe/stripe-go/v82/subscription"
)

type StripeProcessor struct{}

func NewStripeProcessor() PaymentProcessor {
	return &StripeProcessor{}
}

// CreateCustomer creates a customer in Stripe
func (s *StripeProcessor) CreateCustomer(ctx context.Context, userId uuid.UUID, email string) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Metadata: map[string]string{
			"user_id": userId.String(),
		},
	}

	cust, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create customer: %w", err)
	}

	return cust.ID, nil
}

// SetupSubscription creates a subscription product in Stripe
func (s *StripeProcessor) SetupSubscription(ctx context.Context, req *SetupSubscriptionReq) (*SetupSubscriptionResp, error) {
	// Create product
	prod, err := product.New(&stripe.ProductParams{
		Name:        stripe.String(req.Name),
		Description: stripe.String(req.Description),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Create recurring price
	subscriptionPrice, err := price.New(&stripe.PriceParams{
		Currency: stripe.String("usd"),
		Product:  stripe.String(prod.ID),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String("month"),
		},
		UnitAmount: stripe.Int64(req.Price),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription price: %w", err)
	}

	// Set default price
	_, err = product.Update(prod.ID, &stripe.ProductParams{
		DefaultPrice: stripe.String(subscriptionPrice.ID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update product default price: %w", err)
	}

	return &SetupSubscriptionResp{
		ProductID: prod.ID,
		PriceID:   subscriptionPrice.ID,
	}, nil
}

// SubscribeToProduct subscribes a customer to a product
func (s *StripeProcessor) SubscribeToProduct(ctx context.Context, req *SubscribeRequest) (*SubscribeResponse, error) {
	// Get product with expanded default_price
	productParams := &stripe.ProductParams{}
	productParams.AddExpand("default_price")

	prod, err := product.Get(req.ProductID, productParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if prod.DefaultPrice == nil {
		return nil, fmt.Errorf("product has no default price")
	}

	if prod.DefaultPrice.Recurring == nil {
		return nil, fmt.Errorf("product %s is not a subscription", req.ProductID)
	}

	// Create subscription
	subParams := &stripe.SubscriptionParams{
		Customer: stripe.String(req.CustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(prod.DefaultPrice.ID),
			},
		},
		PaymentBehavior: stripe.String("default_incomplete"),
		PaymentSettings: &stripe.SubscriptionPaymentSettingsParams{
			SaveDefaultPaymentMethod: stripe.String("on_subscription"),
		},
	}
	subParams.AddExpand("latest_invoice.confirmation_secret")

	sub, err := subscription.New(subParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Extract client secret
	var clientSecret string
	if sub.LatestInvoice != nil && sub.LatestInvoice.ConfirmationSecret != nil {
		clientSecret = sub.LatestInvoice.ConfirmationSecret.ClientSecret
	}

	return &SubscribeResponse{
		SubscriptionID: sub.ID,
		ClientSecret:   clientSecret,
		Status:         string(sub.Status),
	}, nil
}

// getStripeCustomer retrieves a Stripe customer by ID (used for metadata lookup).
func getStripeCustomer(customerID string) (*stripe.Customer, error) {
	cust, err := customer.Get(customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer %s: %w", customerID, err)
	}
	return cust, nil
}

// GetUserSubscriptions returns all subscriptions for a customer
func (s *StripeProcessor) GetUserSubscriptions(ctx context.Context, customerID string) (*UserSubscriptionsResponse, error) {
	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
	}
	params.AddExpand("data.items.data.price")

	iter := subscription.List(params)

	var subscriptions []UserSubscriptionInfo
	for iter.Next() {
		sub := iter.Subscription()

		info := UserSubscriptionInfo{
			SubscriptionID: sub.ID,
			Status:         string(sub.Status),
		}

		// Get product, price and period info from first item
		if len(sub.Items.Data) > 0 {
			item := sub.Items.Data[0]
			info.PriceID = item.Price.ID
			info.CurrentPeriodEnd = item.CurrentPeriodEnd
			if item.Price.Product != nil {
				info.ProductID = item.Price.Product.ID
			}
		}

		subscriptions = append(subscriptions, info)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("error listing subscriptions: %w", err)
	}

	return &UserSubscriptionsResponse{
		Subscriptions: subscriptions,
	}, nil
}
