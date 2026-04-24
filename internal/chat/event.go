package chat

import "encoding/json"

type Event struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func NewEvent(eventType string, payload any) []byte {
	data, _ := json.Marshal(Event{Type: eventType, Payload: payload})
	return data
}
