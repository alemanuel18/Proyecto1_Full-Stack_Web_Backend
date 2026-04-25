package config

import (
	"log"
	"os"
)

type Config struct {
	DatabaseURL      string
	JWTSecret        string
	CloudinaryURL     string
	CloudinaryCloud  string
	CloudinaryKey    string
	CloudinarySecret string
	Port             string
}

func Load() *Config {
	cfg := &Config{
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production"),
		CloudinaryURL:    getEnv("CLOUDINARY_URL", ""),
		CloudinaryCloud:  getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryKey:    getEnv("CLOUDINARY_API_KEY", ""),
		CloudinarySecret: getEnv("CLOUDINARY_API_SECRET", ""),
		Port:             getEnv("PORT", "8080"),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if cfg.CloudinaryCloud == "" {
		log.Fatal("CLOUDINARY_CLOUD_NAME is required")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}