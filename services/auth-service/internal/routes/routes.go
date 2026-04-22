package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruthwikkakumani/url-shortener/services/auth-service/internal/handler"
	"github.com/ruthwikkakumani/url-shortener/services/auth-service/internal/repository"
	"github.com/ruthwikkakumani/url-shortener/services/auth-service/internal/service"
	"go.uber.org/zap"
)

func RegisterRoutes(r *gin.Engine, logger *zap.Logger, pool *pgxpool.Pool) {
	
	userRepo := repository.NewUserRepo(logger, pool)
	authService := service.NewAuthService(logger, userRepo)
	authHandler := handler.NewAuthHandler(logger, authService)
	
	r.POST("/api/register", authHandler.RegisterHandler)
	
	r.POST("/api/login", authHandler.LoginHandler)
}