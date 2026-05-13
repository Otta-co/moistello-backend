package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/moistello/backend/internal/domain/invite"
)

type Repository struct {
	mock.Mock
}

func (m *Repository) FindByCode(ctx context.Context, code string) (*invite.Invite, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*invite.Invite), args.Error(1)
}

func (m *Repository) FindByCircle(ctx context.Context, circleID uuid.UUID) ([]invite.Invite, error) {
	args := m.Called(ctx, circleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]invite.Invite), args.Error(1)
}

func (m *Repository) Create(ctx context.Context, i *invite.Invite) error {
	return m.Called(ctx, i).Error(0)
}

func (m *Repository) IncrementUse(ctx context.Context, code string) error {
	return m.Called(ctx, code).Error(0)
}

func (m *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
