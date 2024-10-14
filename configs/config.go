package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	DBHost     string `mapstructure:"db_host"`
	DBPort     string `mapstructure:"db_port"`
	DBUser     string `mapstructure:"db_user"`
	DBPassword string `mapstructure:"db_password"`
	DBName     string `mapstructure:"db_name"`
	ServerPort string `mapstructure:"server_port"`
}

func LoadConfig(logger *zap.Logger) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AutomaticEnv()

	// Log current working directory
	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get current working directory", zap.Error(err))
	} else {
		logger.Info("Current working directory", zap.String("cwd", cwd))
	}

	// Attempt to find the config file
	configFile := viper.ConfigFileUsed()
	if configFile != "" {
		absPath, err := filepath.Abs(configFile)
		if err != nil {
			logger.Error("Failed to get absolute path of config file", zap.Error(err))
		} else {
			logger.Info("Config file found", zap.String("path", absPath))
		}
	} else {
		logger.Error("No config file found")
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	// Log the loaded configuration
	logger.Info("Configuration loaded",
		zap.String("db_host", config.DBHost),
		zap.String("db_port", config.DBPort),
		zap.String("db_user", config.DBUser),
		zap.String("db_name", config.DBName),
		zap.String("server_port", config.ServerPort))

	// Validate required fields
	if config.DBHost == "" || config.DBPort == "" || config.DBUser == "" || config.DBName == "" || config.ServerPort == "" {
		return nil, fmt.Errorf("missing required configuration")
	}

	return &config, nil
}
