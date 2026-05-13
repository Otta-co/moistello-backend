package contribution

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Contribution, error)
	FindByCircleAndUser(ctx context.Context, circleID, userID uuid.UUID) (*Contribution, error)
	Create(ctx context.Context, c *Contribution) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status ContributionStatus, txnHash string) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]Contribution, int, error)
	ListByCircle(ctx context.Context, circleID uuid.UUID, page, limit int) ([]Contribution, int, error)
}
