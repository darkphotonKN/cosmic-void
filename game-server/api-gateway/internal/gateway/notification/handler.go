package notification

import (
	"net/http"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/notification"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	client NotificationClient
}

func NewHandler(client NotificationClient) *Handler {
	return &Handler{
		client: client,
	}
}

func (h *Handler) GetNotificationsByUserIDHandler(c *gin.Context) {
	// Get the user ID string from context (set by auth middleware)
	userIdStr, exists := c.Get("userIdStr")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"message":    "User ID not found in context",
		})
		return
	}

	// Create the gRPC request
	req := &pb.NotificationRequest{
		UserId: userIdStr.(string),
		// Limit and Offset are optional, can be added from query params if needed
	}

	// Call the notification service
	response, err := h.client.GetNotification(c.Request.Context(), req)
	if err != nil {
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
		case codes.NotFound:
			httpStatus = http.StatusNotFound
		case codes.InvalidArgument:
			httpStatus = http.StatusBadRequest
		}

		c.JSON(httpStatus, gin.H{
			"statusCode": httpStatus,
			"message":    st.Message(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode":    http.StatusOK,
		"message":       "Successfully retrieved notifications",
		"notifications": response.Notifications,
		"total":         response.Total,
	})
}
