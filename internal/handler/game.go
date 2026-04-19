package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AddGamePayload struct {
	IgdbId int32 `json:"igdbId" binding:"required"`
}

func (h *Handler) SearchGames(c *gin.Context) {
	name, ok := c.GetQuery("name")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name not set"})
		return
	}

	results, err := h.games.Search(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (h *Handler) GetGameInfo(c *gin.Context) {
	igdbId, err := strconv.Atoi(c.Param("igdbId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid igdbId"})
		return
	}

	game, err := h.games.GetOrFetch(c.Request.Context(), int32(igdbId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, game)
}

func (h *Handler) RemoveGame(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)

	userID := claims["sub"].(string)
	gameID := c.Param("id")

	if err := h.games.RemoveFromUserLibrary(c.Request.Context(), userID, gameID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "game removed from library"})
}

func (h *Handler) AddGame(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)

	var payload AddGamePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := claims["sub"].(string)
	if err := h.games.AddToUserLibrary(c.Request.Context(), userID, payload.IgdbId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "game added to library"})
}
