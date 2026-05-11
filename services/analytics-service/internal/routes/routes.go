package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/handler"
	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/repository"
	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/service"
	"go.uber.org/zap"
)

func RegisterRoutes(r *gin.Engine, logger *zap.Logger, pool *pgxpool.Pool) {
	repo := repository.NewAnalyticsRepo(pool, logger)
	svc := service.NewAnalyticsService(repo, logger)
	h := handler.NewAnalyticsHandler(svc, logger)

	// The API Gateway strips /api/analytics before forwarding.
	// So we register the routes at the root, starting with /:code.
	analytics := r.Group("/:code")
	{
		analytics.GET("/summary", h.Summary)      // total clicks + unique visitors
		analytics.GET("/over-time", h.OverTime)   // ?interval=hour|day|week
		analytics.GET("/countries", h.Countries)  // by country
		analytics.GET("/cities", h.Cities)        // by city
		analytics.GET("/devices", h.Devices)      // desktop|mobile|tablet
		analytics.GET("/os", h.OS)                // OS breakdown
		analytics.GET("/browsers", h.Browsers)    // browser breakdown
		analytics.GET("/peak-hours", h.PeakHours) // traffic by hour (0–23)
		analytics.GET("/recent", h.RecentClicks)  // ?limit=20
	}

	// Health probe
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
