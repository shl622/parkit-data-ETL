package config

import (
	"os"
	"parkit-data-ETL/internal/database"
	"parkit-data-ETL/internal/nyc"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoDB   *database.Config
	NYCAPI    *nyc.Config
	BatchSize int
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	cfg := &Config{
		MongoDB: &database.Config{
			URI:      os.Getenv("MONGODB_URI"),
			Database: os.Getenv("MONGODB_DATABASE"),
		},
		NYCAPI: &nyc.Config{
			BaseURL: os.Getenv("NYC_API_URL"),
			Key:     os.Getenv("NYC_API_APP_TOKEN"),
		},
	}
	return cfg, nil
}
