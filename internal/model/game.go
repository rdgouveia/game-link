package model

type Game struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Summary  string `json:"summary"`
	ImageUrl string `json:"imageUrl"`
	IgdbID   int32  `json:"igdbId"`
}
