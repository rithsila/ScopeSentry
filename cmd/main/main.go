// @title           ScopeSentry API
// @version         1.0
// @description     ScopeSentry 是一个安全扫描平台
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8082
// @BasePath  /api

// @securityDefinitions.apikey  ApiKeyAuth
// @in                         header
// @name                       Authorization
// @description               Bearer token for authentication

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Autumn-27/ScopeSentry/docs"
	"github.com/Autumn-27/ScopeSentry/internal/api/routes"
	"github.com/Autumn-27/ScopeSentry/internal/bootstrap"
	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/constants"
	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	sasRedis "github.com/Autumn-27/ScopeSentry/internal/database/redis"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	scopesentrymcp "github.com/Autumn-27/ScopeSentry/internal/mcp"
	"github.com/Autumn-27/ScopeSentry/internal/scheduler"
	redisLogSubscriber "github.com/Autumn-27/ScopeSentry/internal/services/redis_log_subscriber"
	"github.com/Autumn-27/ScopeSentry/internal/update"
	"github.com/Autumn-27/ScopeSentry/internal/worker"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func buildRouter() *gin.Engine {
	router := routes.SetupRouter()
	scopesentrymcp.RegisterRoutes(router)

	// Swagger API docs
	docs.SwaggerInfo.BasePath = "/api"
	router.GET("/api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Scan evidence images
	router.Static("/images", config.GlobalConfig.System.ImgDir)

	// Health check endpoints
	router.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		if !mongodb.EnsureConnected() || mongodb.Client == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "mongodb not connected"})
			return
		}
		if err := mongodb.Client.Ping(ctx, nil); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": fmt.Sprintf("mongodb: %v", err)})
			return
		}

		if sasRedis.Client == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "redis client not initialized"})
			return
		}
		if err := sasRedis.Client.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": fmt.Sprintf("redis: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// Headless fallback: unknown routes return JSON 404
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	})

	return router
}

func main() {
	// Check for headless administrative subcommands before starting the server
	if handled := handleAdminCommands(); handled {
		return
	}

	// Set Gin mode
	if config.GlobalConfig.Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	err := os.Setenv("TZ", config.GlobalConfig.System.Timezone)
	if err != nil {
		return
	}
	if constants.Version == "" {
		constants.Version = "dev"
	}
	logger.Info("Starting ScopeSentry", zap.String("version", constants.Version))
	PrintPlugin()
	if mongodb.EnsureConnected() {
		if err := mongodb.CreateDatabase(); err != nil {
			logger.Error("Database initialization error", zap.Error(err))
		}
	}
	err = update.Update()
	if err != nil {
		logger.Error("Update failed", zap.Error(err))
	}
	bootstrap.GetProjectList()

	// Initialize scheduler
	scheduler.InitializeGlobalScheduler()
	scheduler.GetGlobalScheduler().Start()
	defer scheduler.GetGlobalScheduler().Stop()

	// Start Redis log subscriber
	go redisLogSubscriber.SubscribeLogChannel()

	// Create router
	router := buildRouter()
	gin.DisableConsoleColor()

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.GlobalConfig.Server.Port),
		Handler: router,
	}

	// Run background worker
	worker.Run()

	// Start server in background goroutine
	go func() {
		logger.Info("Starting server", zap.Int("port", config.GlobalConfig.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Graceful shutdown with 5s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}

func Banner() {
	banner := "   _____                         _____            _              \n  / ____|                       / ____|          | |             \n | (___   ___ ___  _ __   ___  | (___   ___ _ __ | |_ _ __ _   _ \n  \\___ \\ / __/ _ \\| '_ \\ / _ \\  \\___ \\ / _ \\ '_ \\| __| '__| | | |\n  ____) | (_| (_) | |_) |  __/  ____) |  __/ | | | |_| |  | |_| |\n |_____/ \\___\\___/| .__/ \\___| |_____/ \\___|_| |_|\\__|_|   \\__, |\n                  | |                                       __/ |\n                  |_|                                      |___/ "
	fmt.Println(banner)
}

func PrintPlugin() {
	if os.Getenv("SCOPESENTRY_PRINT_PLUGIN_KEY") != "true" {
		return
	}
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("✨✨✨ IMPORTANT NOTICE: Please review the Plugin Key below ✨✨✨")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("🔑 Plugin Key: %s\n", config.GlobalConfig.System.PluginKey)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("✅ Ensure the Plugin Key is correctly copied!")
}
