package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"game-link/internal/igdb"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Game struct {
	Id      int    `json:"id" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Summary string `json:"summary" binding:"required"`
	Cover   int    `json:"cover" binding:"required"`
}

type Cover struct {
	Id      int    `json:"id" binding:"required"`
	ImageId string `json:"image_id" binding:"required"`
}

type FullGame struct {
	Game
	ImageUrl string
}

func joinInts(ids []int) string {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = strconv.Itoa(id)
	}
	return strings.Join(strs, ",")
}

func FormatApicalypseStringToSearch(name string) string {
	return fmt.Sprintf("fields id, name, summary, cover; search \"%s\"; where version_parent = null;", name)
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

	body, err := igdb.SearchGamesByName(context.Background(), &http.Client{}, token, searchString)

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

	body, err = igdb.GetGameCovers(context.Background(), &http.Client{}, token, coversQueryString)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var covers []Cover
	json.Unmarshal(body, &covers)

	mergedGames := mergeCoversAndGames(covers, games)

	c.JSON(http.StatusOK, mergedGames)
}
