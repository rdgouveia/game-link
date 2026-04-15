package handler

import (
	"game-link/internal/service"
	"game-link/internal/store"
)

type Handler struct {
	store store.Storer
	games service.GameService
}

func New(s store.Storer, games service.GameService) *Handler {
	return &Handler{store: s, games: games}
}
