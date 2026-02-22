package config

import "os"

type Config struct {
	HTTPAddr    string
	PostgresDSN string
}

func getenvOrDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func LoadConfig() Config {
	return Config{
		HTTPAddr: getenvOrDefault("USER_HTTP_ADDR", ":8081"),
		PostgresDSN: getenvOrDefault(
			"USER_POSTGRES_DSN",
			"postgres://userauth:password@localhost:5432/userauth?sslmode=disable",
		),
	}
}
