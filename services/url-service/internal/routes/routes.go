package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruthwikkakumani/url-shortener/services/url-service/internal/handler"
	"github.com/ruthwikkakumani/url-shortener/services/url-service/internal/middleware"
	"github.com/ruthwikkakumani/url-shortener/services/url-service/internal/repository"
	"github.com/ruthwikkakumani/url-shortener/services/url-service/internal/service"
	"go.uber.org/zap"
)

func RegisterRoutes(r *gin.Engine, logger *zap.Logger , db *pgxpool.Pool) {
	
	repo := repository.NewUrlRepo(logger, db)
	service := service.NewUrlService(logger, repo)
	urlHandler := handler.NewUrlHandler(logger, service)
	
	// Protected group
	authRoutes := r.Group("/")
	authRoutes.Use(middleware.AuthMiddleware())
	
	// Shorten Original URL 
	authRoutes.POST("/api/v1/urls", urlHandler.ShortenURL)
	
	// Redirect Short URL
	r.GET("/:code", urlHandler.RedirectURL)
}