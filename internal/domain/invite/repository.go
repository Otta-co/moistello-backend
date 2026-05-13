package invite

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindByCode(ctx context.Context, code string) (*Invite, error)
	FindByCircle(ctx context.Context, circleID uuid.UUID) ([]Invite, error)
	Create(ctx context.Context, i *Invite) error
	IncrementUse(ctx context.Context, code string) error
	Delete(ctx context.Context, id uuid.UUID) error
}
