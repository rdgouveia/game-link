package store

import (
	"context"
	"fmt"

	"game-link/internal/db"
	"game-link/internal/model"

	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Store) GetMatchCandidates(ctx context.Context, userID string) ([]model.MatchCandidate, error) {
	userUUID, err := stringToUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.q.GetMatchCandidates(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	candidates := make([]model.MatchCandidate, len(rows))
	for i, row := range rows {
		candidates[i] = model.MatchCandidate{
			ID:          uuidToString(row.ID),
			Username:    row.Username,
			Fullname:    row.Fullname,
			Biography:   row.Biography.String,
			Verified:    row.Verified,
			AvatarURL:   row.AvatarUrl.String,
			CommonGames: int(row.CommonGames),
			TotalGames:  int(row.TotalGames),
		}
	}
	return candidates, nil
}

type AvailabilityStorer interface {
	GetUserAvailability(ctx context.Context, userID string) ([]model.AvailabilitySlot, error)
	GetAvailabilityForUsers(ctx context.Context, userIDs []string) (map[string][]model.AvailabilitySlot, error)
	SetUserAvailability(ctx context.Context, userID string, slots []model.AvailabilitySlot) error
}

func (s *Store) GetAvailabilityForUsers(ctx context.Context, userIDs []string) (map[string][]model.AvailabilitySlot, error) {
	uuids := make([]pgtype.UUID, len(userIDs))
	for i, id := range userIDs {
		u, err := stringToUUID(id)
		if err != nil {
			return nil, err
		}
		uuids[i] = u
	}
	rows, err := s.q.GetAvailabilityForUsers(ctx, uuids)
	if err != nil {
		return nil, err
	}
	result := make(map[string][]model.AvailabilitySlot, len(userIDs))
	for _, row := range rows {
		id := uuidToString(row.UserID)
		result[id] = append(result[id], model.AvailabilitySlot{
			DayOfWeek: int(row.DayOfWeek),
			StartTime: pgTimeToString(row.StartTime),
			EndTime:   pgTimeToString(row.EndTime),
		})
	}
	return result, nil
}

func (s *Store) GetUserAvailability(ctx context.Context, userID string) ([]model.AvailabilitySlot, error) {
	userUUID, err := stringToUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.q.GetUserAvailability(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	slots := make([]model.AvailabilitySlot, len(rows))
	for i, row := range rows {
		slots[i] = model.AvailabilitySlot{
			DayOfWeek: int(row.DayOfWeek),
			StartTime: pgTimeToString(row.StartTime),
			EndTime:   pgTimeToString(row.EndTime),
		}
	}
	return slots, nil
}

func (s *Store) SetUserAvailability(ctx context.Context, userID string, slots []model.AvailabilitySlot) error {
	userUUID, err := stringToUUID(userID)
	if err != nil {
		return err
	}
	if err := s.q.DeleteUserAvailability(ctx, userUUID); err != nil {
		return err
	}
	for _, slot := range slots {
		start, err := stringToPgTime(slot.StartTime)
		if err != nil {
			return fmt.Errorf("invalid start_time %q: %w", slot.StartTime, err)
		}
		end, err := stringToPgTime(slot.EndTime)
		if err != nil {
			return fmt.Errorf("invalid end_time %q: %w", slot.EndTime, err)
		}
		if err := s.q.InsertAvailabilitySlot(ctx, db.InsertAvailabilitySlotParams{
			UserID:    userUUID,
			DayOfWeek: int16(slot.DayOfWeek),
			StartTime: start,
			EndTime:   end,
		}); err != nil {
			return err
		}
	}
	return nil
}

func pgTimeToString(t pgtype.Time) string {
	h := t.Microseconds / 3_600_000_000
	m := (t.Microseconds % 3_600_000_000) / 60_000_000
	return fmt.Sprintf("%02d:%02d", h, m)
}

func stringToPgTime(s string) (pgtype.Time, error) {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return pgtype.Time{}, fmt.Errorf("expected HH:MM format")
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return pgtype.Time{}, fmt.Errorf("time out of range")
	}
	return pgtype.Time{
		Microseconds: int64(h)*3_600_000_000 + int64(m)*60_000_000,
		Valid:         true,
	}, nil
}
