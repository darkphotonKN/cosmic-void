package config

import (
	"fmt"

	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/article"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/auth"

	// "github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/build"
	authService "github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/auth"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/build"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/class"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/composite"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/example"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/item"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/notification"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/rating"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/skill"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/tag"
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

	// --- NOTIFICATIONS MICROSERVICE ---

	// -- Notification Setup --
	notificationClient := notification.NewClient(registry)
	notificationHandler := notification.NewHandler(notificationClient)

	// -- Notification Routes --
	notificationRoutes := api.Group("/notifications")

	// Public Routes
	// -- none --

	// Private Routes
	notificationRoutes.Use(auth.AuthMiddleware())
	notificationRoutes.GET("", notificationHandler.GetNotificationsByMemberIdHandler)
	notificationRoutes.POST("/read/:notificationId", notificationHandler.ReadNotificationsByMemberIdHandler)

	/*********************
	* LEGACY MONOLITH APIS
	**********************/

	// --- CLASS AND ASCENDANCY ---

	classClient := class.NewClient(registry)
	classHandler := class.NewHandler(classClient)
	// classRepo := class.NewClassRepository(db)
	// classService := class.NewClassService(classRepo)
	// classHandler := class.NewClassHandler(classService)

	classRoutes := api.Group("/class")
	classRoutes.GET("", classHandler.GetClassesAndAscendanciesHandler)

	// --- SKILL ---

	// -- Skill Setup --

	skillClient := skill.NewClient(registry)
	skillHandler := skill.NewHandler(skillClient)
	// skillRepo := skill.NewSkillRepository(db)
	// skillService := skill.NewSkillService(skillRepo)
	// skillHandler := skill.NewSkillHandler(skillService)

	// -- Skill Routes --
	skillRoutes := api.Group("/skill")

	// Public Routes
	skillRoutes.GET("", skillHandler.GetSkillsHandler)

	// Protected Routes
	skillRoutes.Use(auth.AuthMiddleware())
	skillRoutes.POST("", skillHandler.CreateSkillHandler)

	// --- ITEM ---

	// -- Item Setup --

	itemClient := item.NewClient(registry)
	itemHandler := item.NewHandler(itemClient)
	// itemRepo := item.NewItemRepository(DB)
	// itemService := item.NewItemService(itemRepo, skillService)
	// itemHandler := item.NewItemHandler(itemService)

	// -- Item Routes --
	itemRoutes := api.Group("/item")

	// Protected Routes
	itemRoutes.Use(auth.AuthMiddleware())
	itemRoutes.GET("", itemHandler.GetItemsHandler)
	itemRoutes.POST("", itemHandler.CreateItemHandler)
	itemRoutes.PATCH("", itemHandler.UpdateItemHandler)
	itemRoutes.POST("/rare-item", itemHandler.CreateRareItemHandler)
	// itemRoutes.PATCH("/:id", itemHandler.UpdateItemsHandler)
	// itemRoutes.POST("/rare-item", itemHandler.CreateRareItemHandler)

	itemRoutes.GET("/base-items", itemHandler.GetBaseItemsHandler)
	itemRoutes.GET("/item-mods", itemHandler.GetItemModsHandler)
	itemRoutes.GET("/member-rare-item", itemHandler.GetMemberRareItemsHandler)

	// // base-item, items. skills,
	// itemRoutes.GET("/all-data", itemHandler.GetAllDataHandler)

	// --- BUILD ---

	// -- Build Setup --
	buildClient := build.NewClient(registry)
	buildHandler := build.NewHandler(buildClient)
	// buildRepo := build.NewBuildRepository(db)
	// buildService := build.NewBuildService(buildRepo, skillService)
	// buildHandler := build.NewBuildHandler(buildService)

	// -- Build Routes --
	buildRoutes := api.Group("/build")

	// Public Routes
	buildRoutes.GET("/community", buildHandler.GetCommunityBuildsHandler)
	buildRoutes.GET("/community/:id/info", buildHandler.GetBuildInfoByIdHandler)

	// Protected Routes
	protectedBuildRoutes := buildRoutes.Group("")
	protectedBuildRoutes.Use(auth.AuthMiddleware())
	protectedBuildRoutes.GET("", buildHandler.GetBuildsForMemberHandler)
	protectedBuildRoutes.GET("/:id/info", buildHandler.GetBuildInfoForMemberHandler)
	protectedBuildRoutes.GET("/:id/publish", buildHandler.PublishBuildHandler)
	protectedBuildRoutes.POST("", buildHandler.CreateBuildHandler)
	protectedBuildRoutes.PATCH("/:id", buildHandler.UpdateBuildHandler)
	protectedBuildRoutes.POST("/:id/addSkills", buildHandler.AddSkillLinksToBuildHandler)
	protectedBuildRoutes.PATCH(":id/update-set", buildHandler.UpdateItemSetsToBuildHandler)
	protectedBuildRoutes.DELETE("/:id", buildHandler.DeleteBuildForMemberHandler)

	// --- Composite ---

	// -- Composite Setup --
	compositeClient := composite.NewClient(registry)
	compositeHandler := composite.NewHandler(compositeClient)

	compositeRoutes := api.Group("/composite")

	compositeRoutes.GET("/game-data", compositeHandler.GetGameDataHandler)
	// --- TAG ---

	// -- Tag Setup --

	tagClient := tag.NewClient(registry)
	tagHandler := tag.NewHandler(tagClient)
	// tagRepo := tag.NewTagRepository(db)
	// tagService := tag.NewTagService(tagRepo)
	// tagHandler := tag.NewTagHandler(tagService)

	// -- Tag Routes --
	tagRoutes := api.Group("/tag")

	tagRoutes.GET("", tagHandler.GetTagsHandler)
	// Protected Routes
	tagRoutes.Use(auth.AuthMiddleware())
	tagRoutes.POST("", tagHandler.CreateTagHandler)
	tagRoutes.PATCH("/:id", tagHandler.UpdateTagsHandler)

	// --- Article ---

	// -- Article Setup --
	articleRepo := article.NewArticleRepository(db)
	articleService := article.NewArticleService(articleRepo)
	articleHandler := article.NewArticleHandler(articleService)

	// -- Article Routes --
	articleRoutes := api.Group("/article")

	articleRoutes.GET("", articleHandler.GetArticlesHandler)

	// Protected Routes
	articleRoutes.Use(auth.AuthMiddleware())
	articleRoutes.POST("", articleHandler.CreateArticleHandler)
	articleRoutes.PATCH("/:id", articleHandler.UpdateArticlesHandler)

	// -- RATING --

	// --- Rating Setup ---

	// rating no used in build microservice
	ratingClient := rating.NewClient(registry)
	ratingHandler := rating.NewHandler(ratingClient)
	// ratingRepo := rating.NewRatingRepository(db)
	// ratingService := rating.NewRatingService(ratingRepo, buildService)
	// ratingHandler := rating.NewRatingHandler(ratingService)

	ratingRoutes := api.Group("/rating")

	ratingRoutes.Use(auth.AuthMiddleware())
	ratingRoutes.POST("", ratingHandler.CreateRatingByBuildIdHandler)

	return router
}
