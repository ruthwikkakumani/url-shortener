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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/ruthwikkakumani/url-shortener/pkg/logger"
	"github.com/ruthwikkakumani/url-shortener/services/auth-service/internal/config"
	"github.com/ruthwikkakumani/url-shortener/services/auth-service/internal/db"
	"github.com/ruthwikkakumani/url-shortener/services/auth-service/internal/middleware"
	"github.com/ruthwikkakumani/url-shortener/services/auth-service/internal/routes"
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
func newServer(logger *zap.Logger, pool *pgxpool.Pool) (*gin.Engine) {
	server := gin.New()
	
	server.Use(gin.Recovery())
	
	server.Use(middleware.ZapMiddleware(logger))
	
	routes.RegisterRoutes(server ,logger, pool)
	
	return server
}


func startServer(server *gin.Engine, logger *zap.Logger) {
	port := config.GetEnv("PORT", "8080")
	
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

func main(){
	
	// Load environment variables
	LoadEnv()
	
	env := config.GetEnv("ENV", "development")
	
	// Initialize zap logger
	logger, err := logger.InitLogger(env)
	if err != nil{
		panic(err)
	}
	defer logger.Sync()
	
	dbService := db.NewDB(logger)
	if err := dbService.InitDB(context.Background()); err != nil {
		logger.Fatal("failed to initialize db",
			zap.Error(err),
		)
	}
	defer dbService.Close()
	
	pool, err := dbService.GetPool()
	if err != nil {
		logger.Error("db not initialized",
			zap.Error(err),
		)
	}
	
	// setup server
	server := newServer(logger, pool)
	
	// run server
	startServer(server, logger)
	
}