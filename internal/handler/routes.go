package handler

import (
	"game-link/internal/middleware"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/chat/ws", h.ChatWebSocket)

	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	r.GET("/check/email", h.CheckEmail)
	r.GET("/check/username", h.CheckUsername)

	protected := r.Group("/")
	protected.Use(middleware.Auth())
	{
		protected.GET("/user/me", h.Me)
		protected.PATCH("/user/me", h.UpdateProfile)
		protected.POST("/user/me/avatar", h.UploadAvatar)
		protected.GET("/user/me/availability", h.GetAvailability)
		protected.PUT("/user/me/availability", h.SetAvailability)
		protected.GET("/user/search", h.SearchUsers)
		protected.GET("/user/:username", h.GetUserByUsername)

		protected.GET("/game/search", h.SearchGames)
		protected.GET("/game/:igdbId", h.GetGameInfo)
		protected.POST("/game/add", h.AddGame)
		protected.DELETE("/game/:id", h.RemoveGame)

		protected.GET("/match", h.GetMatches)

		protected.GET("/chat", h.GetConversations)
		protected.GET("/chat/:userId", h.GetConversation)
		protected.PATCH("/chat/:userId/read", h.MarkAsRead)

		protected.POST("/link", h.RequestLink)
		protected.POST("/link/:userId/accept", h.AcceptLink)
		protected.DELETE("/link/:userId", h.RemoveLink)
		protected.GET("/link", h.GetMyLinks)
		protected.GET("/link/pending", h.GetPendingRequests)
	}
}
