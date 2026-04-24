package store

import (
	"context"
	"time"

	"game-link/internal/db"
	"game-link/internal/model"

	"github.com/jackc/pgx/v5/pgtype"
)

type MessageStorer interface {
	InsertMessage(ctx context.Context, senderID, receiverID, content string) (*model.Message, error)
	GetConversation(ctx context.Context, userID1, userID2 string, limit, offset int) ([]model.Message, error)
	GetConversations(ctx context.Context, userID string) ([]model.Conversation, error)
	MarkMessagesAsRead(ctx context.Context, receiverID, senderID string) error
}

func (s *Store) InsertMessage(ctx context.Context, senderID, receiverID, content string) (*model.Message, error) {
	senderUUID, err := stringToUUID(senderID)
	if err != nil {
		return nil, err
	}
	receiverUUID, err := stringToUUID(receiverID)
	if err != nil {
		return nil, err
	}
	row, err := s.q.InsertMessage(ctx, db.InsertMessageParams{
		SenderID:   senderUUID,
		ReceiverID: receiverUUID,
		Content:    content,
	})
	if err != nil {
		return nil, err
	}
	return pgMessageToModel(row.ID, row.SenderID, row.ReceiverID, row.Content, row.ReadAt, row.CreatedAt), nil
}

func (s *Store) GetConversation(ctx context.Context, userID1, userID2 string, limit, offset int) ([]model.Message, error) {
	uuid1, err := stringToUUID(userID1)
	if err != nil {
		return nil, err
	}
	uuid2, err := stringToUUID(userID2)
	if err != nil {
		return nil, err
	}
	rows, err := s.q.GetConversation(ctx, db.GetConversationParams{
		SenderID:   uuid1,
		ReceiverID: uuid2,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, err
	}
	msgs := make([]model.Message, len(rows))
	for i, row := range rows {
		msgs[i] = *pgMessageToModel(row.ID, row.SenderID, row.ReceiverID, row.Content, row.ReadAt, row.CreatedAt)
	}
	return msgs, nil
}

func (s *Store) GetConversations(ctx context.Context, userID string) ([]model.Conversation, error) {
	userUUID, err := stringToUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.q.GetConversations(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	convs := make([]model.Conversation, len(rows))
	for i, row := range rows {
		convs[i] = model.Conversation{
			UserID:      pgInterfaceToString(row.OtherID),
			Username:    row.Username,
			Fullname:    row.Fullname,
			AvatarURL:   row.AvatarUrl.String,
			LastMessage: *pgMessageToModel(row.ID, row.SenderID, row.ReceiverID, row.Content, row.ReadAt, row.CreatedAt),
			UnreadCount: int(row.UnreadCount),
		}
	}
	return convs, nil
}

func (s *Store) MarkMessagesAsRead(ctx context.Context, receiverID, senderID string) error {
	receiverUUID, err := stringToUUID(receiverID)
	if err != nil {
		return err
	}
	senderUUID, err := stringToUUID(senderID)
	if err != nil {
		return err
	}
	return s.q.MarkMessagesAsRead(ctx, db.MarkMessagesAsReadParams{
		ReceiverID: receiverUUID,
		SenderID:   senderUUID,
	})
}

func pgMessageToModel(id, senderID, receiverID pgtype.UUID, content string, readAt, createdAt pgtype.Timestamptz) *model.Message {
	msg := &model.Message{
		ID:         uuidToString(id),
		SenderID:   uuidToString(senderID),
		ReceiverID: uuidToString(receiverID),
		Content:    content,
		CreatedAt:  createdAt.Time,
	}
	if readAt.Valid {
		t := readAt.Time
		msg.ReadAt = &t
	}
	return msg
}

// pgInterfaceToString converte o resultado do CASE WHEN do sqlc (interface{}) para string.
func pgInterfaceToString(v interface{}) string {
	switch val := v.(type) {
	case pgtype.UUID:
		return uuidToString(val)
	case [16]byte:
		var u pgtype.UUID
		u.Bytes = val
		u.Valid = true
		return uuidToString(u)
	case string:
		return val
	default:
		if t, ok := v.(interface{ Bytes() [16]byte }); ok {
			var u pgtype.UUID
			u.Bytes = t.Bytes()
			u.Valid = true
			return uuidToString(u)
		}
		return ""
	}
}

// Garantir que time.Time é usado (evita import não utilizado)
var _ = time.Time{}
