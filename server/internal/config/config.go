package config

import (
	"os"
	"strconv"
	"strings"
)

const (
	defaultCORSOrigin = "http://localhost:5173"
	defaultPort       = 5000
)

type Config struct {
	Port                   int
	NodeEnv                string
	CORSOrigins            []string
	DatabaseURL            string
	ClerkPublishableKey    string
	ClerkSecretKey         string
	ClerkAuthorizedParties []string
	YouTubeAPIKey          string
}

func Load() Config {
	LoadEnvFile(".env")

	origins := splitList(value("CORS_ORIGIN", defaultCORSOrigin))

	parties := splitList(os.Getenv("CLERK_AUTHORIZED_PARTIES"))
	if len(parties) == 0 {
		parties = origins
	}

	nodeEnv := os.Getenv("NODE_ENV")
	if nodeEnv == "" {
		nodeEnv = "development"
	}

	return Config{
		Port:                   parsePort(os.Getenv("PORT")),
		NodeEnv:                nodeEnv,
		CORSOrigins:            origins,
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		ClerkPublishableKey:    os.Getenv("CLERK_PUBLISHABLE_KEY"),
		ClerkSecretKey:         os.Getenv("CLERK_SECRET_KEY"),
		ClerkAuthorizedParties: parties,
		YouTubeAPIKey:          os.Getenv("YOUTUBE_API_KEY"),
	}
}

func IsProduction(cfg Config) bool {
	return cfg.NodeEnv == "production"
}

func value(key string, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parsePort(raw string) int {
	port, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || port <= 0 {
		return defaultPort
	}
	return port
}

func splitList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
