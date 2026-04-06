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
	h := handler.New(s)

	r := gin.Default()
	h.RegisterRoutes(r)

	r.Run(":8080")
}
