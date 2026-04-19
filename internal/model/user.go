package model

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	Fullname  string `json:"fullname"`
	Biography string `json:"biography"`
	Verified  bool   `json:"verified"`
	AvatarURL string `json:"avatarUrl"`
}
