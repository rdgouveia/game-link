package handler

import (
	"net/http"

	"game-link/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func (h *Handler) GetAvailability(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := claims["sub"].(string)

	slots, err := h.store.GetUserAvailability(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, slots)
}

func (h *Handler) SetAvailability(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := claims["sub"].(string)

	var body struct {
		Slots []model.AvailabilitySlot `json:"slots" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, slot := range body.Slots {
		if slot.DayOfWeek < 0 || slot.DayOfWeek > 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "dayOfWeek must be between 0 and 6"})
			return
		}
	}

	if err := h.store.SetUserAvailability(c.Request.Context(), userID, body.Slots); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "availability updated"})
}
