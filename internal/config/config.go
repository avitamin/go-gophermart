// Package config provides application configuration.
package config

import (
	"flag"
	"os"
	"time"
)

// Config holds application configuration.
type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	JWTSecret            string
	TokenDuration        time.Duration
	LogLevel             string
	ShutdownTimeout      time.Duration
	WorkerCount          int
	WorkerInterval       time.Duration
}

// New creates a new Config from environment variables and command-line flags.
func New() *Config {
	cfg := &Config{
		TokenDuration:   7 * 24 * time.Hour, // 7 days
		ShutdownTimeout: 5 * time.Second,
		WorkerCount:     5,
		WorkerInterval:  1 * time.Second,
		LogLevel:        "info",
	}

	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "server address and port")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database connection URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "accrual system address")
	flag.Parse()

	if envAddr := os.Getenv("RUN_ADDRESS"); envAddr != "" {
		cfg.RunAddress = envAddr
	}
	if envDB := os.Getenv("DATABASE_URI"); envDB != "" {
		cfg.DatabaseURI = envDB
	}
	if envAccrual := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrual != "" {
		cfg.AccrualSystemAddress = envAccrual
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "gophermart-secret-key-change-in-production"
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}

	return cfg
}
