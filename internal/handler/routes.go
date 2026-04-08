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
		protected.GET("/user/me", h.Me)
		protected.GET("/game/search", h.SearchGames)
		protected.GET("game/:igdbId", h.GetGameInfo)
		protected.POST("/game/add", h.AddGame)
	}
}
