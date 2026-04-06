package handler

import "game-link/internal/store"

type Handler struct {
	store store.Storer
}

func New(s store.Storer) *Handler {
	return &Handler{store: s}
}
