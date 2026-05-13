package reputation

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	CalculateScore(ctx context.Context, userID string, consecutiveStreaks int, completedCircles int, totalVolumeUSD float64, daysSinceLast int) (*ReputationSnapshot, error)
	UpdateScore(ctx context.Context, userID string, consecutiveStreaks int, completedCircles int, totalVolumeUSD float64, daysSinceLast int) (*ReputationSnapshot, error)
}

type reputationService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &reputationService{repo: repo}
}

func parseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID: %w", err)
	}
	return id, nil
}

func (s *reputationService) CalculateScore(ctx context.Context, userID string, consecutiveStreaks int, completedCircles int, totalVolumeUSD float64, daysSinceLast int) (*ReputationSnapshot, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}
	_ = uid

	streaks := math.Min(float64(consecutiveStreaks)*35, 350)
	completions := math.Min(float64(completedCircles)*50, 300)
	volume := math.Min(math.Log(math.Max(totalVolumeUSD, 1))*30, 200)
	recency := math.Max(0, 150-float64(daysSinceLast)*5)

	total := int(math.Round(streaks + completions + volume + recency))
	if total > 1000 {
		total = 1000
	}

	level := calcLevel(total)

	breakdown := ScoreBreakdown{
		Streaks:     int(streaks),
		Completions: int(completions),
		Volume:      int(volume),
		Recency:     int(recency),
	}

	breakdownJSON, err := json.Marshal(breakdown)
	if err != nil {
		return nil, fmt.Errorf("marshaling breakdown: %w", err)
	}

	return &ReputationSnapshot{
		UserID:    uid,
		Score:     total,
		Level:     level,
		Breakdown: breakdownJSON,
		Month:     time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (s *reputationService) UpdateScore(ctx context.Context, userID string, consecutiveStreaks int, completedCircles int, totalVolumeUSD float64, daysSinceLast int) (*ReputationSnapshot, error) {
	snapshot, err := s.CalculateScore(ctx, userID, consecutiveStreaks, completedCircles, totalVolumeUSD, daysSinceLast)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveSnapshot(ctx, snapshot); err != nil {
		return nil, fmt.Errorf("saving reputation snapshot: %w", err)
	}

	return snapshot, nil
}

func calcLevel(score int) string {
	switch {
	case score > 800:
		return "Diamond"
	case score > 600:
		return "Platinum"
	case score > 400:
		return "Gold"
	case score > 200:
		return "Silver"
	default:
		return "Bronze"
	}
}
