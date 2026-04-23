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
	"github.com/ruthwikkakumani/url-shortener/services/url-service/internal/config"
	"github.com/ruthwikkakumani/url-shortener/services/url-service/internal/middleware"
	"go.uber.org/zap"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment.")
	}
}

func newServer(logger *zap.Logger) (*gin.Engine){
	server := gin.New()
	
	server.Use(gin.Recovery())
	
	server.Use(middleware.ZapMiddleware(logger))
	
	return server
}

func startServer(server *gin.Engine, logger *zap.Logger) {
	port := config.GetEnv("PORT", "8082")
	
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
		
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("server failed to start",
				zap.Error(err),
			)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	
	<-quit
	
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	
	logger.Info("shutting down server....")
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced Shutdown", 
			zap.Error(err),
		)
	}
	
	logger.Info("server exited cleanly")
}

func main() {
	
	// Load env
	LoadEnv()
	
	env := config.GetEnv("ENV", "development")
	
	// Initialize Logger
	logger, err := logger.InitLogger(env)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	
	// server setup
	server := newServer(logger)
	
	// start server
	startServer(server, logger)
}