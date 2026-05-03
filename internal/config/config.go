package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr        string
	PublicBaseURL   string
	DBPath          string
	RawDir          string
	RefreshInterval time.Duration
	BootstrapETL    bool
}

func FromEnv() Config {
	return Config{
		HTTPAddr:        getenv("GUNOI_HTTP_ADDR", ":8080"),
		PublicBaseURL:   getenv("GUNOI_PUBLIC_BASE_URL", "http://localhost:26453"),
		DBPath:          getenv("GUNOI_DB_PATH", "data/db/gunoiarad.db"),
		RawDir:          getenv("GUNOI_RAW_DIR", "data/raw"),
		RefreshInterval: envDuration("GUNOI_REFRESH_INTERVAL", 6*time.Hour),
		BootstrapETL:    envBool("GUNOI_BOOTSTRAP_ETL", false),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
