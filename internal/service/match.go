package service

import (
	"context"
	"fmt"
	"sort"

	"game-link/internal/model"
	"game-link/internal/store"
)

const (
	weightGames    = 0.6
	weightSchedule = 0.4
)

type MatchService interface {
	GetMatches(ctx context.Context, userID string) ([]model.MatchResult, error)
}

type matchService struct {
	store store.Storer
}

func NewMatchService(s store.Storer) MatchService {
	return &matchService{store: s}
}

func (s *matchService) GetMatches(ctx context.Context, userID string) ([]model.MatchResult, error) {
	candidates, err := s.store.GetMatchCandidates(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return []model.MatchResult{}, nil
	}

	myGames, err := s.store.GetUserGames(ctx, userID)
	if err != nil {
		return nil, err
	}

	myAvailability, err := s.store.GetUserAvailability(ctx, userID)
	if err != nil {
		return nil, err
	}
	myTotalMinutes := totalMinutes(myAvailability)

	candidateIDs := make([]string, len(candidates))
	for i, c := range candidates {
		candidateIDs[i] = c.ID
	}
	availabilityMap, err := s.store.GetAvailabilityForUsers(ctx, candidateIDs)
	if err != nil {
		return nil, err
	}

	results := make([]model.MatchResult, 0, len(candidates))
	for _, c := range candidates {
		theirAvailability := availabilityMap[c.ID]

		overlapMinutes := overlapMinutes(myAvailability, theirAvailability)

		gamesScore := gamesScore(c.CommonGames, len(myGames), c.TotalGames)
		scheduleScore := scheduleScore(overlapMinutes, myTotalMinutes)
		score := gamesScore*weightGames + scheduleScore*weightSchedule

		results = append(results, model.MatchResult{
			ID:                     c.ID,
			Username:               c.Username,
			Fullname:               c.Fullname,
			Biography:              c.Biography,
			Verified:               c.Verified,
			AvatarURL:              c.AvatarURL,
			Score:                  roundScore(score),
			CommonGames:            c.CommonGames,
			ScheduleOverlapMinutes: overlapMinutes,
			Availability:           theirAvailability,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

func gamesScore(common, myTotal, theirTotal int) float64 {
	maxGames := max(myTotal, theirTotal)
	if maxGames == 0 {
		return 0
	}
	return float64(common) / float64(maxGames)
}

func scheduleScore(overlapMinutes, myTotalMinutes int) float64 {
	if myTotalMinutes == 0 {
		return 0
	}
	score := float64(overlapMinutes) / float64(myTotalMinutes)
	if score > 1 {
		return 1
	}
	return score
}

func overlapMinutes(a, b []model.AvailabilitySlot) int {
	total := 0
	for _, sa := range a {
		for _, sb := range b {
			if sa.DayOfWeek != sb.DayOfWeek {
				continue
			}
			aStart, aEnd := parseMinutes(sa.StartTime), parseMinutes(sa.EndTime)
			bStart, bEnd := parseMinutes(sb.StartTime), parseMinutes(sb.EndTime)
			start := max(aStart, bStart)
			end := min(aEnd, bEnd)
			if end > start {
				total += end - start
			}
		}
	}
	return total
}

func totalMinutes(slots []model.AvailabilitySlot) int {
	total := 0
	for _, s := range slots {
		total += parseMinutes(s.EndTime) - parseMinutes(s.StartTime)
	}
	return total
}

func parseMinutes(t string) int {
	var h, m int
	fmt.Sscanf(t, "%d:%d", &h, &m)
	return h*60 + m
}

func roundScore(score float64) float64 {
	return float64(int(score*100)) / 100
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
