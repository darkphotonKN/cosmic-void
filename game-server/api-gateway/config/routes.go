package config

import (
	"fmt"

	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/auth"
	authService "github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/auth"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/example"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/item"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/notification"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/payment"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/stats"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

/**
* Sets up API prefix route and all routers.
**/
func SetupRouter(registry discovery.Registry, ch *amqp.Channel) *gin.Engine {
	router := gin.Default()

	// NOTE: debugging middleware
	router.Use(func(c *gin.Context) {
		fmt.Println("Incoming request to:", c.Request.Method, c.Request.URL.Path, "from", c.Request.Host)
		c.Next()
	})

	// TODO: CORS for development, remove in PROD
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	router.Use(otelgin.Middleware("api-gateway"))

	// base route
	api := router.Group("/api")

	/***************
	* MICROSERVICES
	***************/

	// --- EXAMPLE MICROSERVICE ---

	exampleClient := example.NewClient(registry)
	exampleHandler := example.NewHandler(exampleClient)

	exampleRoutes := api.Group("/example")
	exampleRoutes.GET("/:id", exampleHandler.GetExample)
	exampleRoutes.POST("", exampleHandler.CreateExample)

	// --- AUTH & MEMBERS MICROSERVICE ---

	// -- Member Setup (gRPC - for private routes) --
	authClient := authService.NewClient(registry)
	authHandler := authService.NewHandler(authClient)

	// -- Member Setup (AMQP RPC - for signup/signin) --
	amqpAuthHandler := authService.NewAmqpAuthHandler(ch)

	// -- Member Routes --
	memberRoutes := api.Group("/member")

	// Public Routes (via RabbitMQ RPC)
	memberRoutes.POST("/signup", amqpAuthHandler.CreateMemberAmqpHandler)
	memberRoutes.POST("/signin", amqpAuthHandler.LoginMemberAmqpHandler)

	// Private Routes (still via gRPC)
	memberRoutes.Use(auth.AuthMiddleware())
	memberRoutes.GET("", authHandler.GetMemberByIdHandler)
	memberRoutes.PATCH("/update-password", authHandler.UpdatePasswordMemberHandler)
	memberRoutes.PATCH("/update-info", authHandler.UpdateInfoMemberHandler)

	// Avatar Upload Routes (authenticated)
	memberRoutes.POST("/avatar/upload-request", authHandler.RequestAvatarUploadHandler)
	memberRoutes.POST("/avatar/confirm", authHandler.ConfirmAvatarUploadHandler)

	// --- STATS MICROSERVICE ---

	statsClient := stats.NewClient(registry)
	statsHandler := stats.NewHandler(statsClient)

	statsRoutes := api.Group("/stats")
	statsRoutes.GET("/player/:playerId", statsHandler.GetPlayerStats)
	statsRoutes.GET("/leaderboard", statsHandler.GetLeaderboard)

	// --- GAME SERVICE ---
	// TODO: Add game service routes when implemented
	// gameClient := game.NewClient(registry)
	// gameHandler := game.NewHandler(gameClient)
	// gameRoutes := api.Group("/game")
	// gameRoutes.GET("/items", gameHandler.GetItemsHandler)

	// --- NOTIFICATION MICROSERVICE ---

	notificationClient := notification.NewClient(registry)
	notificationHandler := notification.NewHandler(notificationClient)
	notificationRoutes := api.Group("/notification")
	notificationRoutes.Use(auth.AuthMiddleware())
	notificationRoutes.GET("/", notificationHandler.GetNotificationsByUserIDHandler)

	// --- PAYMENT MICROSERVICE ---

	paymentClient := payment.NewClient(registry)
	paymentHandler := payment.NewHandler(paymentClient)

	paymentRoutes := api.Group("/payment")
	paymentRoutes.Use(auth.AuthMiddleware())
	paymentRoutes.POST("/customer", paymentHandler.CreateCustomerHandler)
	paymentRoutes.POST("/subscription/setup", paymentHandler.SetupSubscriptionHandler)
	paymentRoutes.POST("/subscribe", paymentHandler.SubscribeHandler)
	paymentRoutes.GET("/subscriptions/:customerId", paymentHandler.GetUserSubscriptionsHandler)
	paymentRoutes.GET("/subscription/permission", paymentHandler.CheckPermissionHandler)

	// Stripe Webhook (no auth - Stripe sends POST directly)
	router.POST("/webhook/stripe", paymentHandler.WebhookHandler)

	// --- ITEMS MICROSERVICE ---

	itemClient := item.NewClient(registry)
	itemHandler := item.NewHandler(itemClient)

	itemRoutes := api.Group("/items")
	// Private Routes - require authentication
	itemRoutes.Use(auth.AuthMiddleware())

	// --- Legacy/Advanced APIs (creates weapon/armor/consumable separately) ---
	itemRoutes.POST("/weapon", itemHandler.CreateWeaponHandler)
	itemRoutes.POST("/template", itemHandler.CreateItemTemplateHandler) // Creates template only (sends notification)

	// Complete item operations (creates both specific item + template, sends notification)
	itemRoutes.POST("/complete-weapon", itemHandler.CreateCompleteWeaponHandler)
	itemRoutes.POST("/complete-armor", itemHandler.CreateCompleteArmorHandler)
	itemRoutes.POST("/complete-consumable", itemHandler.CreateCompleteConsumableHandler)

	// --- Query APIs ---
	itemRoutes.GET("/weapons", itemHandler.ListWeaponsWithTemplateHandler)

	// --- Dropdown Options (for frontend forms) ---
	itemRoutes.GET("/types", itemHandler.ListItemTypesHandler)
	itemRoutes.GET("/rarities", itemHandler.ListItemRaritiesHandler)

	return router
}
