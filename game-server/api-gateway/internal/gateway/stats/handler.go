package stats

import (
	"net/http"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	client StatsClient
}

func NewHandler(client StatsClient) *Handler {
	return &Handler{
		client: client,
	}
}

func (h *Handler) GetPlayerStats(c *gin.Context) {
	playerID := c.Param("playerId")
	if playerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player ID is required"})
		return
	}

	stats, err := h.client.GetPlayerStats(c.Request.Context(), &pb.GetPlayerStatsRequest{
		PlayerId: playerID,
	})

	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "player stats not found"})
				return
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player ID"})
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, stats)
}