package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruthwikkakumani/redirection-engine/services/auth-service/internal/handler"
	"github.com/ruthwikkakumani/redirection-engine/services/auth-service/internal/repository"
	"github.com/ruthwikkakumani/redirection-engine/services/auth-service/internal/service"
	"go.uber.org/zap"
)

func RegisterRoutes(r *gin.Engine, logger *zap.Logger, pool *pgxpool.Pool) {
	
	userRepo := repository.NewUserRepo(logger, pool)
	authService := service.NewAuthService(logger, userRepo)
	authHandler := handler.NewAuthHandler(logger, authService)
	
	r.POST("/register", authHandler.RegisterHandler)
	r.POST("/login", authHandler.LoginHandler)
	r.POST("/forgot-password", authHandler.ForgotPasswordHandler)
	r.POST("/reset-password", authHandler.ResetPasswordHandler)
}