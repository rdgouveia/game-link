package store

import (
	"errors"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type Store struct {
	mu    sync.RWMutex
	users map[string]*User // keyed by email
}

func New() *Store {
	return &Store{users: make(map[string]*User)}
}

func (s *Store) Create(username, email, password string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[email]; exists {
		return nil, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &User{
		ID:       email,
		Username: username,
		Email:    email,
		Password: string(hash),
	}
	s.users[email] = u
	return u, nil
}

func (s *Store) Authenticate(email, password string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.users[email]
	if !ok {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}
