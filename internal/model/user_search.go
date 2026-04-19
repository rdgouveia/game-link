package model

// UserSearchResult é retornado nas buscas de usuário.
// LinkStatus indica a relação do usuário logado com o resultado:
//
//	"none"             — sem conexão
//	"pending_sent"     — o usuário logado enviou um pedido pendente
//	"pending_received" — o usuário logado recebeu um pedido pendente
//	"accepted"         — já são links
type UserSearchResult struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Fullname   string `json:"fullname"`
	Biography  string `json:"biography"`
	Verified   bool   `json:"verified"`
	AvatarURL  string `json:"avatarUrl"`
	LinkStatus string `json:"linkStatus"`
}
