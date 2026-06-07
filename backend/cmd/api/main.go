package main

import (
	"fmt"
	"os"

	"github.com/drayauction/auctionxi/internal/api/handlers"
	"github.com/drayauction/auctionxi/internal/api/router"
	"github.com/drayauction/auctionxi/internal/repositories"
	"github.com/drayauction/auctionxi/internal/services"
	"github.com/drayauction/auctionxi/pkg/auth"
	"github.com/drayauction/auctionxi/pkg/config"
	"github.com/drayauction/auctionxi/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	log, err := logger.New(cfg.Server.Environment)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	jwtManager := auth.NewJWTManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)

	userRepo := repositories.NewMemoryUserRepository()
	tokenRepo := repositories.NewMemoryRefreshTokenRepository()
	authService := services.NewAuthService(userRepo, tokenRepo, jwtManager)
	auctionService := services.NewAuctionService()

	r := router.New(router.Dependencies{
		AuthHandler:    handlers.NewAuthHandler(authService),
		AuctionHandler: handlers.NewAuctionHandler(auctionService),
		JWTManager:     jwtManager,
		AllowOrigins:   cfg.Server.AllowOrigins,
	})

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Info("starting AuctionXI API", zap.String("addr", addr), zap.String("env", cfg.Server.Environment))
	if err := r.Run(addr); err != nil {
		log.Fatal("server failed", zap.Error(err))
	}
}
