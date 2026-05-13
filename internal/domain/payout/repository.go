package payout

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Payout, error)
	Create(ctx context.Context, p *Payout) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]Payout, int, error)
	ListByCircle(ctx context.Context, circleID uuid.UUID, page, limit int) ([]Payout, int, error)
}
