package stats

import (
	"net/http"
	"strconv"

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

	stats, err := h.client.GetPlayerStats(c.Request.Context(), &pb.GetPlayerMatchStatsRequest{
		MemberId: playerID,
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

func (h *Handler) GetLeaderboard(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	l, err := strconv.Atoi(limitStr)
	if err != nil || l > 50 || l < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 0 and 50."})
		return
	}
	limit := int32(l)

	o, err := strconv.Atoi(offsetStr)
	if err != nil || o < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be 0 or greater."})
		return
	}
	offset := int32(o)

	res, err := h.client.GetLeaderboard(c.Request.Context(), &pb.GetLeaderboardRequest{
		Limit:  &limit,
		Offset: &offset,
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

	c.JSON(http.StatusOK, res)
}
