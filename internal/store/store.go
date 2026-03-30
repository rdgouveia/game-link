package store

import (
	"database/sql"
	"errors"
	"strings"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(username, email, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var u User
	err = s.db.QueryRow(
		`INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id, username, email`,
		username, email, string(hash),
	).Scan(&u.ID, &u.Username, &u.Email)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, errors.New("email already registered")
		}
		return nil, err
	}

	return &u, nil
}

func (s *Store) Authenticate(email, password string) (*User, error) {
	var u User
	err := s.db.QueryRow(
		`SELECT id, username, email, password FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password)
	if err == sql.ErrNoRows {
		return nil, errors.New("invalid credentials")
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return &u, nil
}
