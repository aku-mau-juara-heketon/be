package main

import (
	"fmt"
	"hackathon-be/internal/adapters/handler"
	"hackathon-be/internal/adapters/repository"
	"hackathon-be/internal/core/domain"
	"hackathon-be/internal/core/services"
	"hackathon-be/pkg/config"
	"hackathon-be/pkg/database"
	"hackathon-be/pkg/logger"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Init Logger
	logger.InitLogger()

	// 2. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		return
	}

	// 3. Connect to Database
	db, err := database.Connect(cfg.DBUrl)
	if err != nil {
		return
	}

	// Auto Migrate (For Hackathon speed)
	err = db.AutoMigrate(&domain.User{})
	if err != nil {
		slog.Error("Failed to migrate database", "error", err)
		return
	}

	// 4. Wire Dependencies
	userRepo := repository.NewPostgresUserRepository(db)
	authService := services.NewAuthService(userRepo, cfg)
	authHandler := handler.NewAuthHandler(authService)

	// 5. Setup Router
	r := gin.Default()

	// Public Routes
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/refresh", authHandler.Refresh)
	r.GET("/auth/:provider/login", authHandler.OAuthLogin)
	r.GET("/auth/:provider/callback", authHandler.OAuthCallback)

	// Protected Routes (Example)
	protected := r.Group("/api")
	protected.Use(handler.AuthMiddleware(cfg.JWTSecret))
	{
		protected.GET("/me", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			c.JSON(200, gin.H{"user_id": userID, "message": "You are authenticated"})
		})
	}

	// 6. Start Server
	slog.Info("Starting server", "port", cfg.Port)
	if err := r.Run(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
