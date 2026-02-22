package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr    string
	PostgresDSN string
	JWTSecret   string
}

func getenvOrDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func mustGetenv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("missing required env var: " + key)
	}
	return v
}

func LoadConfig() Config {
	return Config{
		HTTPAddr:    getenvOrDefault("AUTH_HTTP_ADDR", ":8080"),
		PostgresDSN: getenvOrDefault("AUTH_POSTGRES_DSN", "postgres://userauth:password@localhost:5432/userauth?sslmode=disable"),
		JWTSecret:   mustGetenv("JWT_SECRET"),
	}
}

func getenvIntOrDefault(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
