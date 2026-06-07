package router

import (
	"time"

	"github.com/drayauction/auctionxi/internal/api/handlers"
	"github.com/drayauction/auctionxi/internal/api/middleware"
	"github.com/drayauction/auctionxi/internal/repositories"
	"github.com/drayauction/auctionxi/internal/services"
	"github.com/drayauction/auctionxi/pkg/auth"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	AuthHandler    *handlers.AuthHandler
	AuctionHandler *handlers.AuctionHandler
	JWTManager     *auth.JWTManager
	AllowOrigins   []string
}

func New(deps Dependencies) *gin.Engine {
	if gin.Mode() == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(deps.AllowOrigins))
	r.Use(middleware.RateLimit(100, time.Minute))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "auctionxi-api"})
	})

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", deps.AuthHandler.Register)
			auth.POST("/login", deps.AuthHandler.Login)
			auth.POST("/refresh", deps.AuthHandler.Refresh)
			auth.POST("/logout", middleware.AuthMiddleware(deps.JWTManager), deps.AuthHandler.Logout)
		}

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(deps.JWTManager))
		{
			protected.GET("/users/me", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "profile endpoint — wire to user service"})
			})

			auctions := protected.Group("/auctions")
			{
				auctions.POST("", deps.AuctionHandler.Create)
				auctions.GET("/:id", deps.AuctionHandler.Get)
				auctions.POST("/:id/start", deps.AuctionHandler.Start)
				auctions.POST("/:id/bids", deps.AuctionHandler.Bid)
				auctions.POST("/:id/pass", deps.AuctionHandler.Pass)
				auctions.GET("/:id/events", deps.AuctionHandler.Events)
				auctions.GET("/:id/results", deps.AuctionHandler.Results)
			}

			protected.GET("/leaderboards", func(c *gin.Context) {
				c.JSON(200, gin.H{"entries": []interface{}{}})
			})
		}
	}

	return r
}

func NewDefault() *gin.Engine {
	jwtManager := auth.NewJWTManager("dev-access-secret", "dev-refresh-secret", 15*time.Minute, 168*time.Hour)
	userRepo := repositories.NewMemoryUserRepository()
	tokenRepo := repositories.NewMemoryRefreshTokenRepository()
	authService := services.NewAuthService(userRepo, tokenRepo, jwtManager)
	auctionService := services.NewAuctionService()

	return New(Dependencies{
		AuthHandler:    handlers.NewAuthHandler(authService),
		AuctionHandler: handlers.NewAuctionHandler(auctionService),
		JWTManager:     jwtManager,
		AllowOrigins:   []string{"http://localhost:3000"},
	})
}
