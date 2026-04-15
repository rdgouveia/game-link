package store

import (
	"context"
	"errors"
	"fmt"

	"game-link/internal/db"
	"game-link/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

const errCodeUniqueViolation = "23505"

var (
	ErrNotFound    = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Storer interface {
	NewUser(ctx context.Context, username, email, password, fullname string) (*model.User, error)
	Authenticate(ctx context.Context, email, password string) (*model.User, error)
	NewGame(ctx context.Context, name, summary, imageUrl string, igdbID int32) (*model.Game, error)
	GetGameByIgdbID(ctx context.Context, igdbID int32) (*model.Game, error)
	NewUserGame(ctx context.Context, userID, gameID string) error
}

type Store struct {
	q *db.Queries
}

func New(q *db.Queries) *Store {
	return &Store{q: q}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == errCodeUniqueViolation
}

func (s *Store) NewUser(ctx context.Context, username, email, password, fullname string) (*model.User, error) {
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
		if isUniqueViolation(err) {
			return nil, errors.New("email already registered")
		}
		return nil, err
	}

	return &model.User{
		ID:       uuidToString(row.ID),
		Username: row.Username,
		Email:    row.Email,
		Fullname: row.Fullname,
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

func (s *Store) NewGame(ctx context.Context, name, summary, imageUrl string, igdbID int32) (*model.Game, error) {
	row, err := s.q.CreateGame(ctx, db.CreateGameParams{
		Name:     name,
		Summary:  summary,
		ImageUrl: imageUrl,
		IgdbID:   igdbID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}

	return &model.Game{
		ID:       uuidToString(row.ID),
		Name:     row.Name,
		Summary:  row.Summary,
		ImageUrl: row.ImageUrl,
		IgdbID:   row.IgdbID,
	}, nil
}

func (s *Store) GetGameByIgdbID(ctx context.Context, igdbID int32) (*model.Game, error) {
	row, err := s.q.GetGameByIgdbID(ctx, igdbID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &model.Game{
		ID:       uuidToString(row.ID),
		Name:     row.Name,
		Summary:  row.Summary,
		ImageUrl: row.ImageUrl,
		IgdbID:   row.IgdbID,
	}, nil
}

func (s *Store) NewUserGame(ctx context.Context, userID, gameID string) error {
	userUUID, err := stringToUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	gameUUID, err := stringToUUID(gameID)
	if err != nil {
		return fmt.Errorf("invalid game ID: %w", err)
	}

	return s.q.AddGameToUser(ctx, db.AddGameToUserParams{
		UserID: userUUID,
		GameID: gameUUID,
	})
}

func uuidToString(u pgtype.UUID) string {
	b := u.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func stringToUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}
