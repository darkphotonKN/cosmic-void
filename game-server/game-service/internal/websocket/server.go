package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func StartServer(port string, hub *Hub) {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS middleware for WebSocket connections
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// WebSocket endpoint
	router.GET("/ws", func(c *gin.Context) {
		handleWebSocket(hub, c)
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "game-websocket",
		})
	})

	// Stats endpoint
	router.GET("/stats", func(c *gin.Context) {
		roomCount := len(hub.rooms)
		clientCount := len(hub.clients)
		c.JSON(http.StatusOK, gin.H{
			"total_clients": clientCount,
			"total_rooms":   roomCount,
		})
	})

	log.Printf("WebSocket server starting on port %s", port)
	router.Run(":" + port)
}

func handleWebSocket(hub *Hub, c *gin.Context) {
	// Extract user and room info from query params
	userIDStr := c.Query("user_id")
	roomIDStr := c.Query("room_id")

	if userIDStr == "" || roomIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id and room_id are required query parameters",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user_id format",
		})
		return
	}

	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid room_id format",
		})
		return
	}

	// Upgrade the HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Create new client
	client := NewClient(hub, conn, userID, roomID)

	// Register client with hub
	hub.register <- client

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}