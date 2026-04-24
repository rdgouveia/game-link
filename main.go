package main

import (
	"context"
	"log"

	"game-link/internal/chat"
	"game-link/internal/config"
	"game-link/internal/db"
	"game-link/internal/handler"
	"game-link/internal/igdb"
	"game-link/internal/service"
	"game-link/internal/storage"
	"game-link/internal/store"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	godotenv.Load()

	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	if cfg.RedisURL == "" {
		log.Fatal("REDIS_URL environment variable is required")
	}

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal("failed to parse REDIS_URL:", err)
	}

	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("failed to connect to Redis:", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("failed to ping database:", err)
	}

	q := db.New(pool)
	s := store.New(q)
	ig := igdb.NewClient(cfg.IGDBClientID, cfg.IGDBClientSecret, rdb)
	games := service.NewGameService(s, ig)
	links := service.NewLinkService(s)
	match := service.NewMatchService(s)
	hub := chat.NewHub()
	r2 := storage.NewR2Client(cfg.R2AccountID, cfg.R2AccessKeyID, cfg.R2SecretKey, cfg.R2Bucket, cfg.R2PublicURL, cfg.R2AvatarDir)
	h := handler.New(s, games, links, match, hub, r2)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	h.RegisterRoutes(r)

	r.Run(":8080")
}
