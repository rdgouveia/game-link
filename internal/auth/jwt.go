package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"game-link/internal/store"

	"github.com/golang-jwt/jwt/v5"
)

var JWTSecret = []byte("super-secret-key")

func GenerateToken(u *store.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":      u.ID,
		"username": u.Username,
		"email":    u.Email,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(JWTSecret)
}

func ParseToken(tokenStr string) (jwt.MapClaims, error) {
	t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return JWTSecret, nil
	})
	if err != nil || !t.Valid {
		return nil, errors.New("invalid token")
	}
	return t.Claims.(jwt.MapClaims), nil
}

func BearerToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", errors.New("missing bearer token")
	}
	return strings.TrimPrefix(auth, "Bearer "), nil
}
