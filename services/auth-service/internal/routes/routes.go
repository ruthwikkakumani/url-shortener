package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/ruthwikkakumani/redirection-engine/services/auth-service/docs"
	"github.com/ruthwikkakumani/redirection-engine/services/auth-service/internal/config"
	"github.com/ruthwikkakumani/redirection-engine/services/auth-service/internal/handler"
	"github.com/ruthwikkakumani/redirection-engine/services/auth-service/internal/repository"
	"github.com/ruthwikkakumani/redirection-engine/services/auth-service/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func RegisterRoutes(r *gin.Engine, logger *zap.Logger, pool *pgxpool.Pool) {

	userRepo := repository.NewUserRepo(logger, pool)
	authService := service.NewAuthService(logger, userRepo)
	authHandler := handler.NewAuthHandler(logger, authService)

	// Swagger documentation - Only exposed in non-production environments
	if config.GetEnv("ENV", "development") != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	r.POST("/register", authHandler.RegisterHandler)
	r.POST("/login", authHandler.LoginHandler)
	r.POST("/forgot-password", authHandler.ForgotPasswordHandler)
	r.POST("/reset-password", authHandler.ResetPasswordHandler)
}