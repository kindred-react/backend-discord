package config

import (
	"os"
)

type BuiltInUser struct {
	Username string
	Email    string
	Password string
}

type Config struct {
	ServerPort     string
	DatabaseURL    string
	JWTSecret      string
	JWTExpireHours int
	BuiltInUsers   []BuiltInUser
}

func Load() *Config {
	builtInUsers := []BuiltInUser{
		{Username: "甲", Email: "jia@test.com", Password: "password123"},
		{Username: "乙", Email: "yi@test.com", Password: "password123"},
		{Username: "丙", Email: "bing@test.com", Password: "password123"},
	}

	return &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/discord?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "discord-secret-key-change-in-production"),
		JWTExpireHours: 24 * 7,
		BuiltInUsers:   builtInUsers,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
