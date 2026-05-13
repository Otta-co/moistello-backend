package payout

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	Record(ctx context.Context, input RecordInput) (*Payout, error)
	GetUserHistory(ctx context.Context, userID string, page, limit int) ([]Payout, int, error)
	GetCircleHistory(ctx context.Context, circleID string, page, limit int) ([]Payout, int, error)
}

type RecordInput struct {
	CircleID    string     `json:"circleId" validate:"required"`
	RecipientID string     `json:"recipientId" validate:"required"`
	RoundNumber int        `json:"roundNumber" validate:"required,gte=1"`
	Amount      float64    `json:"amount" validate:"required,gt=0"`
	FeeAmount   float64    `json:"feeAmount" validate:"gte=0"`
	TxnHash     string     `json:"txnHash"`
	PayoutType  PayoutType `json:"payoutType" validate:"required,oneof=random fixed auction vote"`
}

type payoutService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &payoutService{repo: repo}
}

func parseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID: %w", err)
	}
	return id, nil
}

func (s *payoutService) Record(ctx context.Context, input RecordInput) (*Payout, error) {
	circleID, err := parseUUID(input.CircleID)
	if err != nil {
		return nil, err
	}
	recipientID, err := parseUUID(input.RecipientID)
	if err != nil {
		return nil, err
	}

	var txnHash sql.NullString
	if input.TxnHash != "" {
		txnHash = sql.NullString{String: input.TxnHash, Valid: true}
	}

	p := &Payout{
		ID:          uuid.New(),
		CircleID:    circleID,
		RecipientID: recipientID,
		RoundNumber: input.RoundNumber,
		Amount:      input.Amount,
		FeeAmount:   input.FeeAmount,
		TxnHash:     txnHash,
		PayoutType:  input.PayoutType,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("recording payout: %w", err)
	}
	return p, nil
}

func (s *payoutService) GetUserHistory(ctx context.Context, userID string, page, limit int) ([]Payout, int, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, 0, err
	}
	payouts, total, err := s.repo.ListByUser(ctx, uid, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("getting user payout history: %w", err)
	}
	return payouts, total, nil
}

func (s *payoutService) GetCircleHistory(ctx context.Context, circleID string, page, limit int) ([]Payout, int, error) {
	cid, err := parseUUID(circleID)
	if err != nil {
		return nil, 0, err
	}
	payouts, total, err := s.repo.ListByCircle(ctx, cid, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("getting circle payout history: %w", err)
	}
	return payouts, total, nil
}
