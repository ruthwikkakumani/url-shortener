package routes

import (
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ruthwikkakumani/url-shortener/services/api-gateway/internal/config"
	"github.com/ruthwikkakumani/url-shortener/services/api-gateway/internal/middleware"
	"go.uber.org/zap"
)
func newProxy(target string, logger *zap.Logger) (*httputil.ReverseProxy){
	u, err := url.Parse(target)
	if err != nil {
		logger.Fatal("invalid proxy target", zap.String("target", target))
	}
	return httputil.NewSingleHostReverseProxy(u)
}

func proxyHandler(proxy *httputil.ReverseProxy, prefix string) (gin.HandlerFunc) {
	return func(c *gin.Context) {
		path := strings.TrimPrefix(c.Request.URL.Path, prefix)
		if path == "" {
			path = "/"
		}
		
		c.Request.URL.Path = path
		c.Request.URL.RawPath = path
		
		clientIP := c.ClientIP()
		if prior := c.Request.Header.Get("X-Forwarded-For"); prior != "" {
			clientIP = prior + ", " + clientIP
		}
		c.Request.Header.Set("X-Forwarded-For", clientIP)
		
		c.Request.Header.Set("X-Gateway", "go-api-gateway")
		
		proxy.ServeHTTP(c.Writer, c.Request)
	} 
}

func RegisterRoutes(r *gin.Engine, logger *zap.Logger) {
	
	cfg := config.Load(logger)
	
	authService := cfg.AuthServiceURL
	urlService := cfg.URLServiceURL
	
	authProxy := newProxy(authService, logger)
	urlProxy := newProxy(urlService, logger)
	
	// Public routes
	// Login & Register
	auth := r.Group("/api/auth")
	{
		auth.Any("/*path", proxyHandler(authProxy, "/api/auth"))
	}
	
	// Redirects
	r.GET("/r/:code", proxyHandler(urlProxy, ""))
	
	urls := r.Group("/api/url")
	urls.Use(middleware.AuthMiddleware())
	{
		urls.Any("/*path", proxyHandler(urlProxy, "/api/url"))
	}
}