package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ruthwikkakumani/url-shortener/services/auth-service/internal/handler"
	"go.uber.org/zap"
)

func RegisterRoutes(r *gin.Engine, logger *zap.Logger) {
	
	authHandler := handler.NewAuthHandler(logger)
	
	r.GET("/api/login", authHandler.RegisterHandler)
}