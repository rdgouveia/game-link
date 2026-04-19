package store

import (
	"context"
	"errors"

	"game-link/internal/db"
	"game-link/internal/model"

	"github.com/jackc/pgx/v5"
)

type LinkStorer interface {
	CreateLink(ctx context.Context, requesterID, recipientID string) error
	AcceptLink(ctx context.Context, requesterID, recipientID string) error
	DeleteLink(ctx context.Context, userID1, userID2 string) error
	GetLinkBetweenUsers(ctx context.Context, userID1, userID2 string) (string, error)
	GetPendingLinksReceived(ctx context.Context, userID string) ([]model.PendingRequest, error)
	GetAcceptedLinks(ctx context.Context, userID string) ([]model.LinkedUser, error)
}

func (s *Store) CreateLink(ctx context.Context, requesterID, recipientID string) error {
	requesterUUID, err := stringToUUID(requesterID)
	if err != nil {
		return err
	}
	recipientUUID, err := stringToUUID(recipientID)
	if err != nil {
		return err
	}
	return s.q.CreateLink(ctx, db.CreateLinkParams{
		RequesterID: requesterUUID,
		RecipientID: recipientUUID,
	})
}

func (s *Store) AcceptLink(ctx context.Context, requesterID, recipientID string) error {
	requesterUUID, err := stringToUUID(requesterID)
	if err != nil {
		return err
	}
	recipientUUID, err := stringToUUID(recipientID)
	if err != nil {
		return err
	}
	return s.q.AcceptLink(ctx, db.AcceptLinkParams{
		RequesterID: requesterUUID,
		RecipientID: recipientUUID,
	})
}

func (s *Store) DeleteLink(ctx context.Context, userID1, userID2 string) error {
	uuid1, err := stringToUUID(userID1)
	if err != nil {
		return err
	}
	uuid2, err := stringToUUID(userID2)
	if err != nil {
		return err
	}
	return s.q.DeleteLink(ctx, db.DeleteLinkParams{
		RequesterID: uuid1,
		RecipientID: uuid2,
	})
}

func (s *Store) GetLinkBetweenUsers(ctx context.Context, userID1, userID2 string) (string, error) {
	uuid1, err := stringToUUID(userID1)
	if err != nil {
		return "", err
	}
	uuid2, err := stringToUUID(userID2)
	if err != nil {
		return "", err
	}
	status, err := s.q.GetLinkBetweenUsers(ctx, db.GetLinkBetweenUsersParams{
		RequesterID: uuid1,
		RecipientID: uuid2,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return status, nil
}

func (s *Store) GetPendingLinksReceived(ctx context.Context, userID string) ([]model.PendingRequest, error) {
	userUUID, err := stringToUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.q.GetPendingLinksReceived(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	requests := make([]model.PendingRequest, len(rows))
	for i, row := range rows {
		requests[i] = model.PendingRequest{
			RequesterID: uuidToString(row.RequesterID),
			Username:    row.Username,
			Fullname:    row.Fullname,
			Verified:    row.Verified,
			AvatarURL:   row.AvatarUrl.String,
			CreatedAt:   row.CreatedAt.Time,
		}
	}
	return requests, nil
}

func (s *Store) GetAcceptedLinks(ctx context.Context, userID string) ([]model.LinkedUser, error) {
	userUUID, err := stringToUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.q.GetAcceptedLinks(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	users := make([]model.LinkedUser, len(rows))
	for i, row := range rows {
		// retorna o outro usuário (não o que fez a requisição)
		if uuidToString(row.RequesterID) == userID {
			users[i] = model.LinkedUser{
				ID:        uuidToString(row.RecipientID),
				Username:  row.RecipientUsername,
				Fullname:  row.RecipientFullname,
				Verified:  row.RecipientVerified,
				AvatarURL: row.RecipientAvatarUrl.String,
			}
		} else {
			users[i] = model.LinkedUser{
				ID:        uuidToString(row.RequesterID),
				Username:  row.RequesterUsername,
				Fullname:  row.RequesterFullname,
				Verified:  row.RequesterVerified,
				AvatarURL: row.RequesterAvatarUrl.String,
			}
		}
	}
	return users, nil
}
