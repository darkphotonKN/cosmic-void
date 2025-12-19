package auth

import (
	"fmt"
	"net/http"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	client AuthClient
}

func NewHandler(client AuthClient) *Handler {
	return &Handler{
		client: client,
	}
}

type Signup struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) CreateMemberHandler(c *gin.Context) {

	var req pb.CreateMemberRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Error parsing payload as JSON"})
		return
	}

	member, err := h.client.CreateMember(c.Request.Context(), &req)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"statusCode": http.StatusInternalServerError,
				"message":    "Internal server error",
			})
			return
		}

		httpStatus := http.StatusInternalServerError
		switch status.Code() {
		case codes.InvalidArgument:
			httpStatus = http.StatusBadRequest
		case codes.AlreadyExists:
			httpStatus = http.StatusConflict
		}

		c.JSON(httpStatus, gin.H{
			"statusCode": httpStatus,
			"message":    status.Message(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"message":    "Successfully created user",
		"result":     member,
	})
}

func (h *Handler) LoginMemberHandler(c *gin.Context) {
	var req pb.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error parsing payload as JSON: %s", err)})
		return
	}

	response, err := h.client.LoginMember(c.Request.Context(), &req)

	if err != nil {
		status, ok := status.FromError(err)

		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"statusCode": http.StatusInternalServerError,
				"message":    "Internal server error",
			})
			return
		}

		httpStatus := http.StatusInternalServerError
		switch status.Code() {
		case codes.InvalidArgument:
			httpStatus = http.StatusBadRequest
		case codes.Unauthenticated:
			httpStatus = http.StatusUnauthorized
		case codes.NotFound:
			httpStatus = http.StatusNotFound
		}

		c.JSON(httpStatus, gin.H{
			"statusCode": httpStatus,
			"message":    status.Message(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"message":    "Successfully logged in",
		"result":     response,
	})
}

func (h *Handler) GetMemberByIdHandler(c *gin.Context) {
	// Get the user ID string from context (set by auth middleware)
	userIdStr, exists := c.Get("userIdStr")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"message":    "User ID not found in context",
		})
		return
	}

	// Create the request
	req := &pb.GetMemberRequest{
		Id: userIdStr.(string),
	}

	// Call the service
	member, err := h.client.GetMember(c.Request.Context(), req)

	if err != nil {
		status, ok := status.FromError(err)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"statusCode": http.StatusInternalServerError,
				"message":    "Internal server error",
			})
			return
		}

		httpStatus := http.StatusInternalServerError
		switch status.Code() {
		case codes.NotFound:
			httpStatus = http.StatusNotFound
		case codes.InvalidArgument:
			httpStatus = http.StatusBadRequest
		}

		c.JSON(httpStatus, gin.H{
			"statusCode": httpStatus,
			"message":    status.Message(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"message":    "Successfully retrieved member",
		"result":     member,
	})
}

func (h *Handler) UpdatePasswordMemberHandler(c *gin.Context) {
	var req pb.UpdatePasswordRequest

	// Get the user ID string from context (set by auth middleware)
	userIdStr, exists := c.Get("userIdStr")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"message":    "User ID not found in context",
		})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Error parsing payload as JSON"})
		return
	}

	// Set the ID from context
	req.Id = userIdStr.(string)

	response, err := h.client.UpdateMemberPassword(c.Request.Context(), &req)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"statusCode": http.StatusInternalServerError,
				"message":    "Internal server error",
			})
			return
		}

		httpStatus := http.StatusInternalServerError
		switch status.Code() {
		case codes.InvalidArgument:
			httpStatus = http.StatusBadRequest
		case codes.NotFound:
			httpStatus = http.StatusNotFound
		case codes.Unauthenticated:
			httpStatus = http.StatusUnauthorized
		}

		c.JSON(httpStatus, gin.H{
			"statusCode": httpStatus,
			"message":    status.Message(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"message":    response.Message,
		"success":    response.Success,
	})
}

func (h *Handler) UpdateInfoMemberHandler(c *gin.Context) {
	var req pb.UpdateMemberInfoRequest

	// Get the user ID string from context (set by auth middleware)
	userIdStr, exists := c.Get("userIdStr")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": http.StatusUnauthorized,
			"message":    "User ID not found in context",
		})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Error parsing payload as JSON"})
		return
	}

	// Set the ID from context
	req.Id = userIdStr.(string)

	member, err := h.client.UpdateMemberInfo(c.Request.Context(), &req)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"statusCode": http.StatusInternalServerError,
				"message":    "Internal server error",
			})
			return
		}

		httpStatus := http.StatusInternalServerError
		switch status.Code() {
		case codes.InvalidArgument:
			httpStatus = http.StatusBadRequest
		case codes.NotFound:
			httpStatus = http.StatusNotFound
		}

		c.JSON(httpStatus, gin.H{
			"statusCode": httpStatus,
			"message":    status.Message(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"message":    "Successfully updated member info",
		"result":     member,
	})
}

func (h *Handler) ValidateTokenHandler(c *gin.Context) {
	var req pb.ValidateTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Error parsing payload as JSON"})
		return
	}

	response, err := h.client.ValidateToken(c.Request.Context(), &req)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"statusCode": http.StatusInternalServerError,
				"message":    "Internal server error",
			})
			return
		}

		httpStatus := http.StatusInternalServerError
		switch status.Code() {
		case codes.InvalidArgument:
			httpStatus = http.StatusBadRequest
		case codes.Unauthenticated:
			httpStatus = http.StatusUnauthorized
		}

		c.JSON(httpStatus, gin.H{
			"statusCode": httpStatus,
			"message":    status.Message(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"valid":      response.Valid,
		"memberId":   response.MemberId,
	})
}
