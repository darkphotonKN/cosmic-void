package auth

import (
	"net/http"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	grpcauth "github.com/darkphotonKN/cosmic-void-server/game-service/grpc/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func WSAuthMiddleware(authClient grpcauth.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}

		resp, err := authClient.ValidateToken(c.Request.Context(), &pb.ValidateTokenRequest{
			Token: token,
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to validate token"})
			c.Abort()
			return
		}

		if !resp.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(resp.MemberId)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid member id"})
			c.Abort()
			return
		}

		// 存入 context
		c.Set("userId", userID)
		c.Set("userIdStr", resp.MemberId)

		c.Next()
	}
}
