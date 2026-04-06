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
	Create(ctx context.Context, username, email, password string) (*model.User, error)
	Authenticate(ctx context.Context, email, password string) (*model.User, error)
}

type Store struct {
	q *db.Queries
}

func New(q *db.Queries) *Store {
	return &Store{q: q}
}

func (s *Store) Create(ctx context.Context, username, email, password string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	row, err := s.q.CreateUser(ctx, db.CreateUserParams{
		Username: username,
		Email:    email,
		Password: string(hash),
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
	}, nil
}

func uuidToString(u pgtype.UUID) string {
	b := u.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
