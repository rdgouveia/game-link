package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"game-link/internal/auth"
	"game-link/internal/chat"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type incomingMessage struct {
	To      string `json:"to"`
	Content string `json:"content"`
}

type outgoingMessage struct {
	ID         string `json:"id"`
	SenderID   string `json:"senderId"`
	ReceiverID string `json:"receiverId"`
	Content    string `json:"content"`
	CreatedAt  string `json:"createdAt"`
}

// ChatWebSocket aceita o token via query param (?token=...) porque o browser
// não permite setar headers customizados em conexões WebSocket.
func (h *Handler) ChatWebSocket(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}
	claims, err := auth.ParseToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	userID := claims["sub"].(string)

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	h.hub.Register(userID, ws)
	defer h.hub.Unregister(userID)

	for {
		_, raw, err := ws.ReadMessage()
		if err != nil {
			break
		}

		var msg incomingMessage
		if err := json.Unmarshal(raw, &msg); err != nil || msg.To == "" || msg.Content == "" {
			continue
		}

		// valida se os dois são links antes de salvar
		status, err := h.store.GetLinkBetweenUsers(c.Request.Context(), userID, msg.To)
		if err != nil || status != "accepted" {
			continue
		}

		saved, err := h.store.InsertMessage(c.Request.Context(), userID, msg.To, msg.Content)
		if err != nil {
			continue
		}

		event := chat.NewEvent("message", outgoingMessage{
			ID:         saved.ID,
			SenderID:   saved.SenderID,
			ReceiverID: saved.ReceiverID,
			Content:    saved.Content,
			CreatedAt:  saved.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})

		h.hub.Send(msg.To, event)
		h.hub.Send(userID, event)
	}
}

func (h *Handler) GetConversations(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := claims["sub"].(string)

	convs, err := h.store.GetConversations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, convs)
}

func (h *Handler) GetConversation(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := claims["sub"].(string)
	otherID := c.Param("userId")

	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	msgs, err := h.store.GetConversation(c.Request.Context(), userID, otherID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, msgs)
}

func (h *Handler) MarkAsRead(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := claims["sub"].(string)
	senderID := c.Param("userId")

	if err := h.store.MarkMessagesAsRead(c.Request.Context(), userID, senderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "messages marked as read"})
}
