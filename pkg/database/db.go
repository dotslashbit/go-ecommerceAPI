package database

import (
	"fmt"

	config "github.com/dotslashbit/ecommerce-api/configs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func NewDB(cfg *config.Config, logger *zap.Logger) (*sqlx.DB, error) {
	// Construct the connection string with host and port separated
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	logger.Info("Attempting to connect to database",
		zap.String("host", cfg.DBHost),
		zap.String("port", cfg.DBPort),
		zap.String("user", cfg.DBUser),
		zap.String("dbname", cfg.DBName))

	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		logger.Error("Failed to connect to database",
			zap.Error(err),
			zap.String("connection_string", connectionString))
		return nil, fmt.Errorf("error connecting to db: %w", err)
	}

	if err = db.Ping(); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		return nil, fmt.Errorf("error pinging db: %w", err)
	}

	logger.Info("Successfully connected to database")
	return db, nil
}
