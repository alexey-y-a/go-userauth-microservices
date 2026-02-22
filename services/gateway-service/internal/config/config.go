package config

import "os"

type Config struct {
	HTTPAddr       string
	AuthServiceURL string
	UserServiceURL string
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
		HTTPAddr:       getenvOrDefault("GATEWAY_HTTP_ADDR", ":8082"),
		AuthServiceURL: getenvOrDefault("AUTH_SERVICE_URL", "http://localhost:8080"),
		UserServiceURL: getenvOrDefault("USER_SERVICE_URL", "http://localhost:8081"),
	}
}
