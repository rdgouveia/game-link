package main

import (
	"database/sql"
	"log"
	"os"

	"game-link/internal/handler"
	"game-link/internal/middleware"
	"game-link/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to open database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatal("failed to migrate database:", err)
	}

	s := store.New(db)
	h := handler.New(s)

	r := gin.Default()

	r.POST("/register", h.Register)
	r.POST("/login", h.Login)

	protected := r.Group("/")
	protected.Use(middleware.Auth())
	{
		protected.GET("/me", h.Me)
	}

	r.Run(":8080")
}
