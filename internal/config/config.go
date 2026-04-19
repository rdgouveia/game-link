package config

import (
	"os"
	"strings"
)

type Config struct {
	DatabaseURL      string
	JWTSecret        string
	RedisURL         string
	IGDBClientID     string
	IGDBClientSecret string
	AllowedOrigins   []string
	R2AccountID      string
	R2AccessKeyID    string
	R2SecretKey      string
	R2Bucket         string
	R2PublicURL      string
	R2AvatarDir      string
}

func Load() Config {
	origins := os.Getenv("ALLOWED_ORIGINS")

	var allowedOrigins []string
	if origins == "" {
		allowedOrigins = []string{"*"}
	} else {
		for _, o := range strings.Split(origins, ",") {
			allowedOrigins = append(allowedOrigins, strings.TrimSpace(o))
		}
	}

	return Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		RedisURL:         os.Getenv("REDIS_URL"),
		IGDBClientID:     os.Getenv("IGDB_CLIENT_ID"),
		IGDBClientSecret: os.Getenv("IGDB_CLIENT_SECRET"),
		AllowedOrigins:   allowedOrigins,
		R2AccountID:      os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyID:    os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretKey:      os.Getenv("R2_SECRET_KEY"),
		R2Bucket:         os.Getenv("R2_BUCKET"),
		R2PublicURL:      os.Getenv("R2_PUBLIC_URL"),
		R2AvatarDir:      os.Getenv("R2_AVATAR_DIR"),
	}
}
