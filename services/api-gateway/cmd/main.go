package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/ruthwikkakumani/url-shortener/pkg/logger"
	"github.com/ruthwikkakumani/url-shortener/services/api-gateway/internal/config"
	"github.com/ruthwikkakumani/url-shortener/services/api-gateway/internal/middleware"
	"github.com/ruthwikkakumani/url-shortener/services/api-gateway/routes"
	"go.uber.org/zap"
)

// Load .env
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment")
	}
}

// setup server
func newServer(logger *zap.Logger) (*gin.Engine){
	server := gin.New()
	
	server.Use(gin.Recovery())
	
	server.Use(middleware.ZapMiddleware(logger))
	
	routes.RegisterRoutes(server, logger)
	
	return server
}

// Start server
func startServer(server *gin.Engine, logger *zap.Logger) {
	port := config.GetEnv("PORT", "8081")
	
	srv := &http.Server{
		Addr: ":" + port,
		Handler: server,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 10 * time.Second,
	}
	
	
	go func() {
		logger.Info("Server starting",
			zap.String("port", port),
		)
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed{
			logger.Fatal("server failed to start",
				zap.Error(err),
			)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	
	<-quit
	
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	
	logger.Info("shutting down server...")
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced Shutdown", 
			zap.Error(err),
		)
	}
	
	logger.Info("server exited cleanly")
}

func main() {
	
	// Load environment variables
	LoadEnv();
	
	env := config.GetEnv("ENV", "development")
	
	// Initialize logger
	logger, err := logger.InitLogger(env)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	
	// setup server
	server := newServer(logger)
	
	// run server
	startServer(server, logger)
}