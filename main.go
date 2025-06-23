package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-url-shortener/handler"
	"go-url-shortener/storage"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Use default args for normal execution
	runServer(os.Args)
}

// runServer starts the server with the given command-line arguments
// This function is exported for testing purposes
func runServer(args []string) {
	// Save original args and restore them later
	oldArgs := os.Args
	os.Args = args
	defer func() { os.Args = oldArgs }()
	
	// Parse command-line flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	port := flag.Int("port", 8080, "Port to listen on")
	dbType := flag.String("db", "memory", "Database type (memory or sqlite)")
	dbPath := flag.String("db-path", "urls.db", "Path to SQLite database (only for sqlite)")
	flag.Parse()

	// Configure structured logging
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := config.Build()
	defer logger.Sync()

	// Create the store
	var store storage.Store
	var err error

	switch *dbType {
	case "memory":
		logger.Info("Using in-memory storage")
		store = storage.NewMemoryStore()
	case "sqlite":
		logger.Info("Using SQLite storage", zap.String("path", *dbPath))
		store, err = storage.NewSQLiteStore(*dbPath)
		if err != nil {
			logger.Fatal("Failed to create SQLite store", zap.Error(err))
		}
	default:
		logger.Fatal("Invalid database type", zap.String("type", *dbType))
	}
	defer store.Close()

	// Create a prometheus registry
	registry := prometheus.NewRegistry()

	// Register default collectors
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	registry.MustRegister(prometheus.NewGoCollector())

	// Create handler with store and prometheus registry
	urlHandler := handler.NewURLHandler(store, registry)

	// Create router
	router := gin.New()
	router.Use(ginZapMiddleware(logger))
	router.Use(gin.Recovery())

	// API routes
	router.POST("/api/shorten", urlHandler.Shorten)
	router.GET("/api/stats", urlHandler.GetStats)
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
	router.GET("/:id", urlHandler.Redirect)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server started", zap.Int("port", *port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown gracefully
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// ginZapMiddleware returns a gin middleware that logs requests using zap
func ginZapMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		logger.Info("Request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
		)
	}
}
