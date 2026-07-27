package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
}

// Load reads configuration from the environment. If a .env file is present
// in the working directory it is loaded first (without overriding variables
// already set in the environment); its absence is not an error.
func Load() (*Config, error) {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{DatabaseURL: dsn, Port: port}, nil
}
