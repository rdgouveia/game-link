package igdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const tokenCacheKey = "igdb:access_token"

type TwitchToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func GetAccessToken(rdb *redis.Client) (string, error) {
	ctx := context.Background()

	cached, err := rdb.Get(ctx, tokenCacheKey).Result()
	if err == nil {
		return cached, nil
	}

	url := fmt.Sprintf(
		"https://id.twitch.tv/oauth2/token?client_id=%s&client_secret=%s&grant_type=client_credentials",
		os.Getenv("IGDB_CLIENT_ID"), os.Getenv("IGDB_CLIENT_SECRET"),
	)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var token TwitchToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", err
	}

	ttl := time.Duration(token.ExpiresIn)*time.Second - 5*time.Minute
	rdb.Set(ctx, tokenCacheKey, token.AccessToken, ttl)

	return token.AccessToken, nil
}

func SearchGamesByName(ctx context.Context, client *http.Client, token string, searchString string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.igdb.com/v4/games",
		strings.NewReader(searchString),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Client-ID", os.Getenv("IGDB_CLIENT_ID"))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func GetGameCovers(ctx context.Context, client *http.Client, token string, searchString string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.igdb.com/v4/covers",
		strings.NewReader(searchString),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Client-ID", os.Getenv("IGDB_CLIENT_ID"))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
