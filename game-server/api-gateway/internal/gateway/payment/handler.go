package payment

import (
	"net/http"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/payment"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	client PaymentClient
}

func NewHandler(client PaymentClient) *Handler {
	return &Handler{
		client: client,
	}
}

func (h *Handler) CreateCustomerHandler(c *gin.Context) {
	userIdStr, exists := c.Get("userIdStr")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"message":    "User ID not found in context",
		})
		return
	}

	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"message":    "Error parsing payload as JSON",
		})
		return
	}

	resp, err := h.client.CreateCustomer(c.Request.Context(), &pb.CreateCustomerRequest{
		UserId: userIdStr.(string),
		Email:  req.Email,
	})
	if err != nil {
		handleGrpcError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"message":    "Successfully created customer",
		"result":     resp,
	})
}

func (h *Handler) SetupSubscriptionHandler(c *gin.Context) {
	var req pb.SetupSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"message":    "Error parsing payload as JSON",
		})
		return
	}

	resp, err := h.client.SetupSubscription(c.Request.Context(), &req)
	if err != nil {
		handleGrpcError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"message":    "Successfully setup subscription product",
		"result":     resp,
	})
}

func (h *Handler) SubscribeHandler(c *gin.Context) {
	userIdStr, exists := c.Get("userIdStr")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"message":    "User ID not found in context",
		})
		return
	}

	var req struct {
		ProductID string `json:"product_id" binding:"required"`
		Email     string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"message":    "Error parsing payload as JSON",
		})
		return
	}

	ctx := c.Request.Context()

	// Step 1: Auto-create Stripe customer
	custResp, err := h.client.CreateCustomer(ctx, &pb.CreateCustomerRequest{
		UserId: userIdStr.(string),
		Email:  req.Email,
	})
	if err != nil {
		handleGrpcError(c, err)
		return
	}

	// Step 2: Subscribe with the customer ID
	resp, err := h.client.Subscribe(ctx, &pb.SubscribeRequest{
		ProductId:  req.ProductID,
		CustomerId: custResp.CustomerId,
	})
	if err != nil {
		handleGrpcError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"message":    "Successfully subscribed",
		"result":     resp,
	})
}

func (h *Handler) GetUserSubscriptionsHandler(c *gin.Context) {
	customerID := c.Param("customerId")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"message":    "Customer ID is required",
		})
		return
	}

	resp, err := h.client.GetUserSubscriptions(c.Request.Context(), &pb.GetUserSubscriptionsRequest{
		CustomerId: customerID,
	})
	if err != nil {
		handleGrpcError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"message":    "Successfully retrieved subscriptions",
		"result":     resp,
	})
}

func handleGrpcError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"message":    "Internal server error",
		})
		return
	}

	httpStatus := http.StatusInternalServerError
	switch st.Code() {
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.Unavailable:
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"statusCode": httpStatus,
		"message":    st.Message(),
	})
}
