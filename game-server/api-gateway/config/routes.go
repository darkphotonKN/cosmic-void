package config

import (
	"fmt"

	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/auth"
	authService "github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/auth"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/example"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

/**
* Sets up API prefix route and all routers.
**/
func SetupRouter(registry discovery.Registry, db *sqlx.DB) *gin.Engine {
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

	// -- Member Setup --
	authClient := authService.NewClient(registry)
	authHandler := authService.NewHandler(authClient)

	// -- Member Routes --
	memberRoutes := api.Group("/member")

	// Public Routes
	memberRoutes.POST("/signup", authHandler.CreateMemberHandler)
	memberRoutes.POST("/signin", authHandler.LoginMemberHandler)

	// Private Routes
	memberRoutes.Use(auth.AuthMiddleware())
	memberRoutes.GET("", authHandler.GetMemberByIdHandler)
	memberRoutes.PATCH("/update-password", authHandler.UpdatePasswordMemberHandler)
	memberRoutes.PATCH("/update-info", authHandler.UpdateInfoMemberHandler)

	// --- GAME SERVICE ---
	// TODO: Add game service routes when implemented
	// gameClient := game.NewClient(registry)
	// gameHandler := game.NewHandler(gameClient)
	// gameRoutes := api.Group("/game")
	// gameRoutes.GET("/items", gameHandler.GetItemsHandler)

	return router
}