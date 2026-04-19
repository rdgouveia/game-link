package handler

import (
	"game-link/internal/service"
	"game-link/internal/storage"
	"game-link/internal/store"
)

type Handler struct {
	store store.Storer
	games service.GameService
	links service.LinkService
	r2    *storage.R2Client
}

func New(s store.Storer, games service.GameService, links service.LinkService, r2 *storage.R2Client) *Handler {
	return &Handler{store: s, games: games, links: links, r2: r2}
}
