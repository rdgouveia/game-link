package main

import (
	"game-link/internal/handler"
	"game-link/internal/middleware"
	"game-link/internal/store"

	"github.com/gin-gonic/gin"
)

func main() {
	db := store.New()
	h := handler.New(db)

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
