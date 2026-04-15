package config

import "os"

type Config struct {
	DatabaseURL      string
	JWTSecret        string
	RedisURL         string
	IGDBClientID     string
	IGDBClientSecret string
}

func Load() Config {
	return Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		RedisURL:         os.Getenv("REDIS_URL"),
		IGDBClientID:     os.Getenv("IGDB_CLIENT_ID"),
		IGDBClientSecret: os.Getenv("IGDB_CLIENT_SECRET"),
	}
}
