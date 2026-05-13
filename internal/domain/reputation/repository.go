package reputation

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	GetByUser(ctx context.Context, userID uuid.UUID) (*ReputationSnapshot, error)
	SaveSnapshot(ctx context.Context, s *ReputationSnapshot) error
	GetHistory(ctx context.Context, userID uuid.UUID, months int) ([]ReputationSnapshot, error)
}
