package igdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const tokenCacheKey = "igdb:access_token"

// Client é a interface pública do pacote igdb.
type Client interface {
	SearchGames(ctx context.Context, name string) ([]SearchResult, error)
	GetGameWithCover(ctx context.Context, igdbID int32) (*GameDetail, error)
}

// SearchResult representa um jogo retornado pela busca da IGDB.
type SearchResult struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GameDetail representa os dados completos de um jogo com imagem de capa.
type GameDetail struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Summary  string `json:"summary"`
	ImageURL string `json:"imageUrl"`
}

type igdbClient struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	rdb          *redis.Client
}

// NewClient cria um novo Client da IGDB com as credenciais e Redis fornecidos.
func NewClient(clientID, clientSecret string, rdb *redis.Client) Client {
	return &igdbClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{},
		rdb:          rdb,
	}
}

func (c *igdbClient) getAccessToken() (string, error) {
	ctx := context.Background()

	if cached, err := c.rdb.Get(ctx, tokenCacheKey).Result(); err == nil {
		return cached, nil
	}

	url := fmt.Sprintf(
		"https://id.twitch.tv/oauth2/token?client_id=%s&client_secret=%s&grant_type=client_credentials",
		c.clientID, c.clientSecret,
	)

	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var token struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", err
	}

	ttl := time.Duration(token.ExpiresIn)*time.Second - 5*time.Minute
	c.rdb.Set(ctx, tokenCacheKey, token.AccessToken, ttl)

	return token.AccessToken, nil
}

func (c *igdbClient) post(ctx context.Context, endpoint, body string) ([]byte, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Client-ID", c.clientID)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *igdbClient) SearchGames(ctx context.Context, name string) ([]SearchResult, error) {
	query := fmt.Sprintf(`fields id, name; search "%s"; where version_parent = null;`, name)

	data, err := c.post(ctx, "https://api.igdb.com/v4/games", query)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (c *igdbClient) GetGameWithCover(ctx context.Context, igdbID int32) (*GameDetail, error) {
	query := fmt.Sprintf("fields id, name, summary, cover; where id = %d;", igdbID)

	data, err := c.post(ctx, "https://api.igdb.com/v4/games", query)
	if err != nil {
		return nil, err
	}

	var games []struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Summary string `json:"summary"`
		Cover   int    `json:"cover"`
	}
	if err := json.Unmarshal(data, &games); err != nil {
		return nil, err
	}
	if len(games) == 0 {
		return nil, fmt.Errorf("game %d not found on IGDB", igdbID)
	}

	game := games[0]

	coverQuery := fmt.Sprintf("fields id, image_id; where id = (%d);", game.Cover)
	coverData, err := c.post(ctx, "https://api.igdb.com/v4/covers", coverQuery)
	if err != nil {
		return nil, err
	}

	var covers []struct {
		ID      int    `json:"id"`
		ImageID string `json:"image_id"`
	}
	if err := json.Unmarshal(coverData, &covers); err != nil {
		return nil, err
	}

	var imageURL string
	if len(covers) > 0 {
		imageURL = fmt.Sprintf("https://images.igdb.com/igdb/image/upload/t_1080p_2x/%s.jpg", covers[0].ImageID)
	}

	return &GameDetail{
		ID:       game.ID,
		Name:     game.Name,
		Summary:  game.Summary,
		ImageURL: imageURL,
	}, nil
}
