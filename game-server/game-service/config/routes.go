package config

import (
	"fmt"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/gameserver"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

/**
* Sets up API prefix route and all routers.
**/
func SetupRouter(db *sqlx.DB) *gin.Engine {
	router := gin.Default()

	// NOTE: debugging middleware
	router.Use(func(c *gin.Context) {
		fmt.Println("Incoming request to:", c.Request.Method, c.Request.URL.Path, "from", c.Request.Host)
		c.Next()
	})

	// CORS for development - more specific for game service with WebSocket support
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "http://127.0.0.1:3000", "http://localhost:3838"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Upgrade", "Connection", "Sec-WebSocket-Key", "Sec-WebSocket-Version", "Sec-WebSocket-Extensions"},
		AllowCredentials: true,
	}))

	// base route
	api := router.Group("/api")

	// --- WEBSOCKET CONNECTION ---
	server := gameserver.NewServer()

	// -- routes --
	router.GET("/game/ws", server.HandleWebSocketConnection)

	// --- HEALTH CHECK ---
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "game-service",
		})
	})

	// TODO: Add game specific routes
	// gameRoutes := api.Group("/game")
	// gameRoutes.GET("/items", gameHandler.GetItems)

	return router
}
