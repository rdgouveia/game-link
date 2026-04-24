package chat

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	conns map[string]*conn
	mu    sync.RWMutex
}

type conn struct {
	ws *websocket.Conn
	mu sync.Mutex
}

func NewHub() *Hub {
	return &Hub{conns: make(map[string]*conn)}
}

func (h *Hub) Register(userID string, ws *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[userID] = &conn{ws: ws}
}

func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, userID)
}

func (h *Hub) Send(userID string, data []byte) bool {
	h.mu.RLock()
	c, ok := h.conns[userID]
	h.mu.RUnlock()
	if !ok {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ws.WriteMessage(websocket.TextMessage, data)
	return true
}

func (h *Hub) Online(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.conns[userID]
	return ok
}
