package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"game-link/internal/db"
	"game-link/internal/model"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type Storer interface {
	NewUser(ctx context.Context, username, email, password string, fullname string) (*model.User, error)
	Authenticate(ctx context.Context, email, password string) (*model.User, error)
	NewGame(ctx context.Context, name string, summary string, imageUrl string, igdbID int32) (*model.Game, error)
	GetGameID(ctx context.Context, igdbID int32) (*db.Game, error)
	NewUserGame(ctx context.Context, user_id pgtype.UUID, game_id pgtype.UUID) error
}

type Store struct {
	q *db.Queries
}

func New(q *db.Queries) *Store {
	return &Store{q: q}
}

func (s *Store) NewUser(ctx context.Context, username, email, password string, fullname string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	row, err := s.q.CreateUser(ctx, db.CreateUserParams{
		Username: username,
		Email:    email,
		Password: string(hash),
		Fullname: fullname,
	})

	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, errors.New("email already registered")
		}
		return nil, err
	}

	return &model.User{
		ID:       uuidToString(row.ID),
		Username: row.Username,
		Email:    row.Email,
	}, nil
}

func (s *Store) Authenticate(ctx context.Context, email, password string) (*model.User, error) {
	row, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(row.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &model.User{
		ID:       uuidToString(row.ID),
		Username: row.Username,
		Email:    row.Email,
		Fullname: row.Fullname,
	}, nil
}

func (s *Store) NewGame(ctx context.Context, name string, summary string, imageUrl string, igdbID int32) (*model.Game, error) {
	row, err := s.q.CreateGame(ctx, db.CreateGameParams{
		Name:     name,
		Summary:  summary,
		ImageUrl: imageUrl,
		IgdbID:   igdbID,
	})

	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, errors.New("Error while adding game to database")
		}
		return nil, err
	}

	return &model.Game{
		ID:       uuidToString(row.ID),
		Name:     name,
		Summary:  summary,
		ImageUrl: imageUrl,
		IgdbID:   igdbID,
	}, nil
}

func (s *Store) GetGameID(ctx context.Context, igdbID int32) (*db.Game, error) {
	row, err := s.q.GetGameByIgdbID(ctx, igdbID)

	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, errors.New("Game not found on database")
		}
		return nil, err
	}

	return &row, nil
}

func (s *Store) NewUserGame(ctx context.Context, userId pgtype.UUID, gameId pgtype.UUID) error {
	err := s.q.AddGameToUser(ctx, db.AddGameToUserParams{
		UserID: userId,
		GameID: gameId,
	})

	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return errors.New("Error while adding game to database")
		}
		return err
	}

	return nil
}

func uuidToString(u pgtype.UUID) string {
	b := u.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
