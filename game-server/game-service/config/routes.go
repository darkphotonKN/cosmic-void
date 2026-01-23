package config

import (
	"fmt"

	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/darkphotonKN/cosmic-void-server/game-service/auth"
	grpcauth "github.com/darkphotonKN/cosmic-void-server/game-service/grpc/auth"
	grpcitems "github.com/darkphotonKN/cosmic-void-server/game-service/grpc/items"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/gameserver"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/queue"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"

	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
)

/**
* Sets up API prefix route and all routers.
**/
func SetupRouter(db *sqlx.DB, registry discovery.Registry, ch *amqp.Channel) *gin.Engine {
	router := gin.Default()

	// NOTE: debugging middleware
	router.Use(func(c *gin.Context) {
		fmt.Println("Incoming request to:", c.Request.Method, c.Request.URL.Path, "from", c.Request.Host)
		c.Next()
	})

	// CORS for development more specific for game service with WebSocket support
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "http://127.0.0.1:3000", "http://localhost:3838"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Upgrade", "Connection", "Sec-WebSocket-Key", "Sec-WebSocket-Version", "Sec-WebSocket-Extensions"},
		AllowCredentials: true,
	}))

	// base route
	api := router.Group("/api")

	// --- AUTH CLIENT ---
	authClient := grpcauth.NewClient(registry)
	itemsClient := grpcitems.NewClient(registry)

	// --- GAME SERVER SETUP ---
	queueService := queue.NewQueueService(2)

	// wrap with adapter to allow amqp rabbit mq channel to
	// conform to our abstraction
	publishCh := commonbroker.NewAmqpPublisher(ch)
	gameService := game.NewService(publishCh)
	server := gameserver.NewServer(authClient, queueService, gameService, itemsClient)

	// -- routes --
	router.GET("/game/ws", auth.WSAuthMiddleware(authClient), server.HandleWebSocketConnection)

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
