package handler

import (
	"game-link/internal/middleware"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)

	protected := r.Group("/")
	protected.Use(middleware.Auth())
	{
		protected.GET("/me", h.Me)
	}
}
