package main

import (
	"log"

	config "github.com/dotslashbit/ecommerce-api/configs"
	"github.com/dotslashbit/ecommerce-api/internal/product" // New import
	"github.com/dotslashbit/ecommerce-api/pkg/database"
	"github.com/dotslashbit/ecommerce-api/pkg/server"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment() // Using Development logger for more verbose output
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database
	db, err := database.NewDB(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize product repository
	productRepo := product.NewRepository(db)

	// Initialize product service
	productService := product.NewService(productRepo)

	// Initialize product handler
	productHandler := product.NewHandler(productService, logger)

	// Initialize server
	srv := server.NewServer(db, logger)

	// Register product routes
	productHandler.RegisterRoutes(srv.Router)

	// Start server
	logger.Info("Starting server", zap.String("port", cfg.ServerPort))
	if err := srv.Start(":" + cfg.ServerPort); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}
