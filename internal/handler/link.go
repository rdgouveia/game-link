package handler

import (
	"net/http"

	"game-link/internal/chat"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type linkRequestPayload struct {
	UserID string `json:"userId" binding:"required"`
}

func (h *Handler) RequestLink(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)
	requesterID := claims["sub"].(string)

	var payload linkRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.links.RequestLink(c.Request.Context(), requesterID, payload.UserID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	if user, err := h.store.GetUserByID(c.Request.Context(), requesterID); err == nil {
		h.hub.Send(payload.UserID, chat.NewEvent("link_request", gin.H{
			"fromId":    requesterID,
			"username":  user.Username,
			"fullname":  user.Fullname,
			"avatarUrl": user.AvatarURL,
		}))
	}

	c.JSON(http.StatusCreated, gin.H{"message": "link request sent"})
}

func (h *Handler) AcceptLink(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)
	recipientID := claims["sub"].(string)
	requesterID := c.Param("userId")

	if err := h.links.AcceptLink(c.Request.Context(), recipientID, requesterID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user, err := h.store.GetUserByID(c.Request.Context(), recipientID); err == nil {
		h.hub.Send(requesterID, chat.NewEvent("link_accepted", gin.H{
			"fromId":    recipientID,
			"username":  user.Username,
			"fullname":  user.Fullname,
			"avatarUrl": user.AvatarURL,
		}))
	}

	c.JSON(http.StatusOK, gin.H{"message": "link accepted"})
}

func (h *Handler) RemoveLink(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := claims["sub"].(string)
	otherUserID := c.Param("userId")

	if err := h.links.RemoveLink(c.Request.Context(), userID, otherUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "link removed"})
}

func (h *Handler) GetMyLinks(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := claims["sub"].(string)

	users, err := h.links.GetMyLinks(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *Handler) GetPendingRequests(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := claims["sub"].(string)

	requests, err := h.links.GetPendingRequests(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, requests)
}
