package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

// Config represents the configuration for the application.
type Config struct {
	// Logger is the configuration for the logger.
	Logger LoggerConfig `yaml:"logger"`
	// GRPC is the configuration for the gRPC server.
	GRPC GRPCConfig `yaml:"grpc"`

	// GracefulShutdownTimeout is the timeout for graceful shutdown.
	GracefulShutdownTimeout time.Duration `yaml:"graceful_shutdown_timeout"`
}

// LoggerConfig represents the configuration for the logger.
type LoggerConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// GRPCConfig represents the configuration for the gRPC server.
type GRPCConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

// NewConfig creates a new Config with default values.
func NewConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	config := &Config{}
	if err := yaml.NewDecoder(file).Decode(config); err != nil {
		return nil, err
	}
	if err := config.ValidateAndSetDefault(); err != nil {
		return nil, err
	}
	return config, nil
}

// ValidateAndSetDefault validates the configuration and sets default values if necessary.
func (c *Config) ValidateAndSetDefault() error {
	if c.Logger.Level == "" {
		c.Logger.Level = "info"
	}
	if c.Logger.Format == "" {
		c.Logger.Format = "json"
	}
	if c.GRPC.Port == "" {
		return fmt.Errorf("gRPC port is required")
	}
	return nil
}
