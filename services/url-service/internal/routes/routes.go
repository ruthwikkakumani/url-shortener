package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/cache"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/handler"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/middleware"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/repository"
	"github.com/ruthwikkakumani/redirection-engine/services/url-service/internal/service"
	"go.uber.org/zap"
)

func RegisterRoutes(r *gin.Engine, logger *zap.Logger, db *pgxpool.Pool, cache *cache.RedisClient) {

	repo := repository.NewUrlRepo(logger, db)
	urlService := service.NewUrlService(logger, repo, cache)
	urlHandler := handler.NewUrlHandler(logger, urlService)

	// Protected routes
	urls := r.Group("/")
	protected := urls.Group("")
	protected.Use(middleware.AuthMiddleware())

	// Shorten Original URL
	protected.POST("", urlHandler.ShortenURL)

	// List registered urls
	protected.GET("/urls", urlHandler.ListURLS)

	// Update shorten URL
	protected.PATCH("/:shortCode", urlHandler.UpdateURL)

	// Delete URL
	protected.DELETE("/:shortCode", urlHandler.DeleteURL)

}
