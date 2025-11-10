package router

import (
	"konnect/docs"
	"konnect/internal/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(router *gin.Engine, middleware *handler.Middleware, authHandler *handler.AuthHandler, profileHandler *handler.ProfileHandler) {
	// cors
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "X-Forwarded-For", "Origin", "Content-Type", "Content-Length"},
		AllowCredentials: true,
	}))

	// base api router
	apiRouter := router.Group("/api")
	apiRouter.GET("/health", func(ctx *gin.Context) { ctx.JSON(200, "OK") })

	// swagger
	docs.SwaggerInfo.BasePath = "/api"
	apiRouter.GET("/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// auth routes
	auth := apiRouter.Group("/auth")
	{
		auth.GET("/google/init", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)
	}

	// protected routes
	protected := apiRouter.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// profiles
		profiles := protected.Group("/profiles")
		{
			profiles.POST("", profileHandler.CreateProfile)
			profiles.POST("/photo", profileHandler.UploadProfilePhoto)
			profiles.GET("/:id", profileHandler.GetProfile)
			profiles.PATCH("", profileHandler.UpdateProfile)
			profiles.GET("/nearby", profileHandler.GetNearbyProfiles)
		}
	}
}
