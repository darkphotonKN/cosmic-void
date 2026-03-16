package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	grpcauth "github.com/darkphotonKN/cosmic-void-server/game-service/grpc/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Custom close code for auth errors (4000-4999 range is for application use)
const CloseCodeAuthError = 4001

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// rejectWebSocket upgrades the connection to WebSocket, sends an auth error message,
// then closes with custom close code 4001 so the client can identify auth failures.
func rejectWebSocket(c *gin.Context, reason string) {
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("Failed to upgrade WebSocket for auth rejection: %v\n", err)
		c.Abort()
		return
	}
	defer conn.Close()

	// Send auth_error message so client knows the reason
	authErrMsg, _ := json.Marshal(map[string]interface{}{
		"action": "auth_error",
		"payload": map[string]string{
			"message": reason,
		},
	})
	conn.WriteMessage(websocket.TextMessage, authErrMsg)

	// Close with custom code 4001
	conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(CloseCodeAuthError, reason),
	)

	c.Abort()
}

func WSAuthMiddleware(authClient grpcauth.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			rejectWebSocket(c, "missing token")
			return
		}

		resp, err := authClient.ValidateToken(c.Request.Context(), &pb.ValidateTokenRequest{
			Token: token,
		})
		if err != nil {
			rejectWebSocket(c, "failed to validate token")
			return
		}

		if !resp.Valid {
			rejectWebSocket(c, "invalid or expired token")
			return
		}

		userID, err := uuid.Parse(resp.MemberId)
		if err != nil {
			rejectWebSocket(c, "invalid member id")
			return
		}

		// store in context
		c.Set("userId", userID)
		c.Set("userIdStr", resp.MemberId)

		c.Next()
	}
}
