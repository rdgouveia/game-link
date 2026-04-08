package handler

import (
	"encoding/json"
	"fmt"
	"game-link/internal/igdb"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type GameBase struct {
	Id   int    `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type Game struct {
	GameBase
	Summary string `json:"summary" binding:"required"`
	Cover   int    `json:"cover" binding:"required"`
}

type Cover struct {
	Id      int    `json:"id" binding:"required"`
	ImageId string `json:"image_id" binding:"required"`
}

type FullGame struct {
	Game
	ImageUrl string `json:"imageUrl" binding:"required"`
}

type AddGamePayload struct {
	IgdbId int32 `json:"igdbId" binding:"required"`
}

func joinInts(ids []int) string {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = strconv.Itoa(id)
	}
	return strings.Join(strs, ",")
}

func FormatApicalypseStringToSearch(name string) string {
	return fmt.Sprintf("fields id, name; search \"%s\"; where version_parent = null;", name)
}

func FormatApicalypseStringToGetGame(igdbId string) string {
	return fmt.Sprintf("fields id, name, summary, cover; where id = %s;", igdbId)
}

func FormatApicalypseStringToCovers(gameCovers []int) string {
	return fmt.Sprintf("fields id, image_id; where id = (%s);", joinInts(gameCovers))
}

func mergeCoversAndGames(covers []Cover, games []Game) []FullGame {
	var mergedGames []FullGame

	coverMap := make(map[int]Cover)
	for _, cover := range covers {
		coverMap[cover.Id] = cover
	}

	for _, game := range games {
		if cover, ok := coverMap[game.Cover]; ok {
			mergedGame := new(FullGame)
			mergedGame.Game = game
			mergedGame.ImageUrl = fmt.Sprintf("https://images.igdb.com/igdb/image/upload/t_1080p_2x/%s.jpg", cover.ImageId)

			mergedGames = append(mergedGames, *mergedGame)
		}
	}

	return mergedGames
}

func (h *Handler) SearchGames(c *gin.Context) {
	name, present := c.GetQuery("name")

	if present == false {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name not set"})
		return
	}

	token, err := igdb.GetAccessToken(h.rdb)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	searchString := FormatApicalypseStringToSearch(name)

	body, err := igdb.Games(c.Request.Context(), &http.Client{}, token, searchString)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var games []GameBase
	json.Unmarshal(body, &games)

	c.JSON(http.StatusOK, games)
}

func (h *Handler) GetGameInfo(c *gin.Context) {
	igdbId := c.Param("igdbId")

	if igdbId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "igdbId not set"})
		return
	}

	igdbIdInt, err := strconv.Atoi(igdbId)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, _ := h.store.GetGameID(c.Request.Context(), int32(igdbIdInt))

	if game != nil {
		var fullGame FullGame
		fullGame.ImageUrl = game.ImageUrl
		fullGame.Id = int(game.IgdbID)
		fullGame.Name = game.Name
		fullGame.Summary = game.Summary

		c.JSON(http.StatusOK, fullGame)
		return
	}

	token, err := igdb.GetAccessToken(h.rdb)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	queryString := FormatApicalypseStringToGetGame(igdbId)

	body, err := igdb.Games(c.Request.Context(), &http.Client{}, token, queryString)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var games []Game
	json.Unmarshal(body, &games)

	var gameCovers []int

	for _, game := range games {
		gameCovers = append(gameCovers, game.Cover)
	}

	coversQueryString := FormatApicalypseStringToCovers(gameCovers)

	body, err = igdb.GetGameCovers(c.Request.Context(), &http.Client{}, token, coversQueryString)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var covers []Cover
	json.Unmarshal(body, &covers)

	mergedGames := mergeCoversAndGames(covers, games)

	_, err = h.store.NewGame(c.Request.Context(), mergedGames[0].Name, mergedGames[0].Summary, mergedGames[0].ImageUrl, int32(mergedGames[0].Id))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mergedGames[0])
}

func (h *Handler) AddGame(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)
	var payload AddGamePayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.store.GetGameID(c.Request.Context(), payload.IgdbId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	parsedUUID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	userUUID := pgtype.UUID{Bytes: parsedUUID, Valid: true}
	err = h.store.NewUserGame(c.Request.Context(), userUUID, game.ID)
}
