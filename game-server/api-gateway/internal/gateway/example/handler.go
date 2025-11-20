package example

import (
	"net/http"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/example"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	client ExampleClient
}

func NewHandler(client ExampleClient) *Handler {
	return &Handler{
		client: client,
	}
}

func (h *Handler) CreateExample(c *gin.Context) {
	var request *pb.CreateExampleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	example, err := h.client.CreateExample(c.Request.Context(), request)

	if err != nil {
		status, ok := status.FromError(err)

		if !ok {
			// not a gRPC status error
			c.JSON(http.StatusInternalServerError, gin.H{
				"statusCode": http.StatusInternalServerError,
				"message":    "Internal server error",
			})

			return
		}

		// map grpc error codes to http codes
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

	c.JSON(http.StatusCreated, gin.H{"statusCode": http.StatusOK, "message": "success", "result": example})
}

func (h *Handler) GetExample(c *gin.Context) {
	id := c.Param("id")

	// Convert REST request to gRPC request
	grpcReq := &pb.GetExampleRequest{
		Id: id,
	}

	// Call the service
	example, err := h.client.GetExample(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"statusCode": http.StatusOK, "message": "success", "result": example})
}
