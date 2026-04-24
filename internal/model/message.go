package model

import "time"

type Message struct {
	ID         string     `json:"id"`
	SenderID   string     `json:"senderId"`
	ReceiverID string     `json:"receiverId"`
	Content    string     `json:"content"`
	ReadAt     *time.Time `json:"readAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type Conversation struct {
	UserID      string   `json:"userId"`
	Username    string   `json:"username"`
	Fullname    string   `json:"fullname"`
	AvatarURL   string   `json:"avatarUrl"`
	LastMessage Message  `json:"lastMessage"`
	UnreadCount int      `json:"unreadCount"`
}
