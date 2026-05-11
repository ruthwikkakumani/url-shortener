package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/ruthwikkakumani/url-shortener/pkg/logger"
	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/config"
	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/db"
	kafkaconsumer "github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/kafka"
	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/middleware"
	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/repository"
	"github.com/ruthwikkakumani/url-shortener/services/analytics-service/internal/routes"
	"go.uber.org/zap"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env found, using system environment")
	}
}

func newServer(logger *zap.Logger, dbPool interface{ /* pgxpool */ }) *gin.Engine {
	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(middleware.ZapMiddleware(logger))
	return server
}

func main() {
	LoadEnv()

	env := config.GetEnv("ENV", "production")

	zapLogger, err := logger.InitLogger(env)
	if err != nil {
		panic(err)
	}
	defer zapLogger.Sync() //nolint:errcheck

	dbService := db.NewDB(zapLogger)
	if err := dbService.InitDB(context.Background()); err != nil {
		zapLogger.Fatal("failed to init analytics db", zap.Error(err))
	}
	defer dbService.Close()

	pool, err := dbService.GetPool()
	if err != nil {
		zapLogger.Fatal("analytics db pool unavailable", zap.Error(err))
	}

	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(middleware.ZapMiddleware(zapLogger))
	routes.RegisterRoutes(server, zapLogger, pool)

	port := config.GetEnv("PORT", "8085")

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      server,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	go func() {
		zapLogger.Info("analytics HTTP server starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("analytics server failed", zap.Error(err))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaBrokers := config.GetEnv("KAFKA_BROKERS", "")
	if kafkaBrokers != "" {
		brokers := strings.Split(kafkaBrokers, ",")
		repo := repository.NewAnalyticsRepo(pool, zapLogger)

		consumer, err := kafkaconsumer.NewConsumer(brokers, repo, zapLogger)
		if err != nil {
			zapLogger.Warn("kafka: failed to init consumer (" + err.Error() + ") — click events will not be persisted")
		} else {
			go consumer.Start(ctx)
			zapLogger.Info("kafka consumer goroutine started", zap.Strings("brokers", brokers))
		}
	} else {
		zapLogger.Warn("KAFKA_BROKERS not set — analytics consumer disabled")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	zapLogger.Info("shutting down analytics service...")
	cancel() // stop consumer

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zapLogger.Error("forced HTTP shutdown", zap.Error(err))
	}

	zapLogger.Info("analytics service exited cleanly")
}