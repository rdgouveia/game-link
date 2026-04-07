package handler

import (
	"game-link/internal/store"

	"github.com/redis/go-redis/v9"
)

type Handler struct {
	store store.Storer
	rdb   *redis.Client
}

func New(s store.Storer, rdb *redis.Client) *Handler {
	return &Handler{store: s, rdb: rdb}
}
