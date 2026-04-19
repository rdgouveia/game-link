package model

import "time"

type LinkedUser struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Fullname  string `json:"fullname"`
	Verified  bool   `json:"verified"`
	AvatarURL string `json:"avatarUrl"`
}

type PendingRequest struct {
	RequesterID string    `json:"requesterId"`
	Username    string    `json:"username"`
	Fullname    string    `json:"fullname"`
	Verified    bool      `json:"verified"`
	AvatarURL   string    `json:"avatarUrl"`
	CreatedAt   time.Time `json:"createdAt"`
}
