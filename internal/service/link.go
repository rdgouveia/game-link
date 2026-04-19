package service

import (
	"context"
	"errors"

	"game-link/internal/model"
	"game-link/internal/store"
)

type LinkService interface {
	RequestLink(ctx context.Context, requesterID, recipientID string) error
	AcceptLink(ctx context.Context, recipientID, requesterID string) error
	RemoveLink(ctx context.Context, userID, otherUserID string) error
	GetMyLinks(ctx context.Context, userID string) ([]model.LinkedUser, error)
	GetPendingRequests(ctx context.Context, userID string) ([]model.PendingRequest, error)
}

type linkService struct {
	store store.LinkStorer
}

func NewLinkService(s store.LinkStorer) LinkService {
	return &linkService{store: s}
}

func (s *linkService) RequestLink(ctx context.Context, requesterID, recipientID string) error {
	if requesterID == recipientID {
		return errors.New("cannot link with yourself")
	}

	status, err := s.store.GetLinkBetweenUsers(ctx, requesterID, recipientID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return err
	}

	switch status {
	case "accepted":
		return errors.New("already linked")
	case "pending":
		return errors.New("link request already sent")
	}

	return s.store.CreateLink(ctx, requesterID, recipientID)
}

func (s *linkService) AcceptLink(ctx context.Context, recipientID, requesterID string) error {
	status, err := s.store.GetLinkBetweenUsers(ctx, requesterID, recipientID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return errors.New("link request not found")
		}
		return err
	}
	if status != "pending" {
		return errors.New("link is not pending")
	}

	return s.store.AcceptLink(ctx, requesterID, recipientID)
}

func (s *linkService) RemoveLink(ctx context.Context, userID, otherUserID string) error {
	return s.store.DeleteLink(ctx, userID, otherUserID)
}

func (s *linkService) GetMyLinks(ctx context.Context, userID string) ([]model.LinkedUser, error) {
	return s.store.GetAcceptedLinks(ctx, userID)
}

func (s *linkService) GetPendingRequests(ctx context.Context, userID string) ([]model.PendingRequest, error) {
	return s.store.GetPendingLinksReceived(ctx, userID)
}
