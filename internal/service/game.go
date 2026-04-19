package service

import (
	"context"
	"errors"
	"fmt"

	"game-link/internal/igdb"
	"game-link/internal/model"
	"game-link/internal/store"
)

type GameService interface {
	Search(ctx context.Context, name string) ([]igdb.SearchResult, error)
	GetOrFetch(ctx context.Context, igdbID int32) (*model.Game, error)
	AddToUserLibrary(ctx context.Context, userID string, igdbID int32) error
	RemoveFromUserLibrary(ctx context.Context, userID, gameID string) error
}

type gameService struct {
	store store.Storer
	igdb  igdb.Client
}

func NewGameService(s store.Storer, ig igdb.Client) GameService {
	return &gameService{store: s, igdb: ig}
}

func (s *gameService) Search(ctx context.Context, name string) ([]igdb.SearchResult, error) {
	return s.igdb.SearchGames(ctx, name)
}

// GetOrFetch busca o jogo no banco local; se não existir, busca na IGDB e salva.
func (s *gameService) GetOrFetch(ctx context.Context, igdbID int32) (*model.Game, error) {
	game, err := s.store.GetGameByIgdbID(ctx, igdbID)
	if err == nil {
		return game, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}

	detail, err := s.igdb.GetGameWithCover(ctx, igdbID)
	if err != nil {
		return nil, fmt.Errorf("fetching game from IGDB: %w", err)
	}

	game, err = s.store.NewGame(ctx, detail.Name, detail.Summary, detail.ImageURL, int32(detail.ID))
	if err != nil {
		// Race condition: outra requisição salvou o jogo primeiro — busca do banco.
		if errors.Is(err, store.ErrAlreadyExists) {
			return s.store.GetGameByIgdbID(ctx, igdbID)
		}
		return nil, err
	}

	return game, nil
}

// RemoveFromUserLibrary remove a associação entre o usuário e o jogo pelo ID do banco.
func (s *gameService) RemoveFromUserLibrary(ctx context.Context, userID, gameID string) error {
	return s.store.RemoveUserGame(ctx, userID, gameID)
}

// AddToUserLibrary garante que o jogo existe no banco e o associa ao usuário.
func (s *gameService) AddToUserLibrary(ctx context.Context, userID string, igdbID int32) error {
	game, err := s.GetOrFetch(ctx, igdbID)
	if err != nil {
		return err
	}
	return s.store.NewUserGame(ctx, userID, game.ID)
}
