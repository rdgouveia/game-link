package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"game-link/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type Store struct {
	q *db.Queries
}

func New(q *db.Queries) *Store {
	return &Store{q: q}
}

func (s *Store) Create(username, email, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	row, err := s.q.CreateUser(context.Background(), db.CreateUserParams{
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

	return &User{
		ID:       uuidToString(row.ID),
		Username: row.Username,
		Email:    row.Email,
	}, nil
}

func (s *Store) Authenticate(email, password string) (*User, error) {
	row, err := s.q.GetUserByEmail(context.Background(), email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(row.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &User{
		ID:       uuidToString(row.ID),
		Username: row.Username,
		Email:    row.Email,
	}, nil
}

func uuidToString(u pgtype.UUID) string {
	b := u.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
