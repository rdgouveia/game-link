package model

type MatchCandidate struct {
	ID          string
	Username    string
	Fullname    string
	Biography   string
	Verified    bool
	AvatarURL   string
	CommonGames int
	TotalGames  int
}

type MatchResult struct {
	ID                     string               `json:"id"`
	Username               string               `json:"username"`
	Fullname               string               `json:"fullname"`
	Biography              string               `json:"biography"`
	Verified               bool                 `json:"verified"`
	AvatarURL              string               `json:"avatarUrl"`
	Score                  float64              `json:"score"`
	CommonGames            int                  `json:"commonGames"`
	ScheduleOverlapMinutes int                  `json:"scheduleOverlapMinutes"`
	Availability           []AvailabilitySlot   `json:"availability"`
}
