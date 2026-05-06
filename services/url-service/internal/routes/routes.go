package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruthwikkakumani/url-shortener/services/url-service/internal/handler"
	"github.com/ruthwikkakumani/url-shortener/services/url-service/internal/repository"
	"github.com/ruthwikkakumani/url-shortener/services/url-service/internal/service"
	"go.uber.org/zap"
)

func RegisterRoutes(r *gin.Engine, logger *zap.Logger , db *pgxpool.Pool) {
	
	repo := repository.NewUrlRepo(logger, db)
	urlService := service.NewUrlService(logger, repo)
	urlHandler := handler.NewUrlHandler(logger, urlService)
	
	// Protected routes 
	urls := r.Group("/")
	protected := urls.Group("")
	
	// Shorten Original URL 
	protected.POST("", urlHandler.ShortenURL)

	// Update shorten URL
	protected.PATCH("/:shortCode", urlHandler.UpdateURL)
	
}