package main

import (
	"context"
	"log"
	"os"

	"game-link/internal/db"
	"game-link/internal/handler"
	"game-link/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL environment variable is required")
	}

	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("failed to parse REDIS_URL:", err)
	}

	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("failed to connect to Redis:", err)
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("failed to ping database:", err)
	}

	q := db.New(pool)
	s := store.New(q)
	h := handler.New(s, rdb)

	r := gin.Default()
	h.RegisterRoutes(r)

	r.Run(":8080")
}
