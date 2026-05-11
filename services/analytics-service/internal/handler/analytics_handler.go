package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/service"
	"go.uber.org/zap"
)

type AnalyticsHandler struct {
	svc    *service.AnalyticsService
	logger *zap.Logger
}

func NewAnalyticsHandler(svc *service.AnalyticsService, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc, logger: logger}
}

// respond is a tiny helper to keep handlers DRY.
func (h *AnalyticsHandler) respond(c *gin.Context, data any, err error) {
	if err != nil {
		h.logger.Error("analytics handler error",
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch analytics"})
		return
	}
	if data == nil {
		c.JSON(http.StatusOK, gin.H{"data": []any{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GET /api/analytics/:code/summary
// Returns total clicks and unique visitors.
func (h *AnalyticsHandler) Summary(c *gin.Context) {
	code := c.Param("code")
	data, err := h.svc.Summary(c.Request.Context(), code)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/over-time?interval=hour|day|week
// Returns time-series click data.
func (h *AnalyticsHandler) OverTime(c *gin.Context) {
	code := c.Param("code")
	interval := c.DefaultQuery("interval", "day")
	data, err := h.svc.OverTime(c.Request.Context(), code, interval)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/countries
func (h *AnalyticsHandler) Countries(c *gin.Context) {
	code := c.Param("code")
	data, err := h.svc.Countries(c.Request.Context(), code)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/cities
func (h *AnalyticsHandler) Cities(c *gin.Context) {
	code := c.Param("code")
	data, err := h.svc.Cities(c.Request.Context(), code)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/devices
func (h *AnalyticsHandler) Devices(c *gin.Context) {
	code := c.Param("code")
	data, err := h.svc.Devices(c.Request.Context(), code)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/os
func (h *AnalyticsHandler) OS(c *gin.Context) {
	code := c.Param("code")
	data, err := h.svc.OSBreakdown(c.Request.Context(), code)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/browsers
func (h *AnalyticsHandler) Browsers(c *gin.Context) {
	code := c.Param("code")
	data, err := h.svc.Browsers(c.Request.Context(), code)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/peak-hours
// Returns hourly traffic distribution (0–23).
func (h *AnalyticsHandler) PeakHours(c *gin.Context) {
	code := c.Param("code")
	data, err := h.svc.PeakHours(c.Request.Context(), code)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/recent?limit=20
// Returns recent click events.
func (h *AnalyticsHandler) RecentClicks(c *gin.Context) {
	code := c.Param("code")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	data, err := h.svc.RecentClicks(c.Request.Context(), code, limit)
	h.respond(c, data, err)
}
