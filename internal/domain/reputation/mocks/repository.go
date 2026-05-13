package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/moistello/backend/internal/domain/reputation"
)

type Repository struct {
	mock.Mock
}

func (m *Repository) GetByUser(ctx context.Context, userID uuid.UUID) (*reputation.ReputationSnapshot, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reputation.ReputationSnapshot), args.Error(1)
}

func (m *Repository) SaveSnapshot(ctx context.Context, s *reputation.ReputationSnapshot) error {
	return m.Called(ctx, s).Error(0)
}

func (m *Repository) GetHistory(ctx context.Context, userID uuid.UUID, months int) ([]reputation.ReputationSnapshot, error) {
	args := m.Called(ctx, userID, months)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]reputation.ReputationSnapshot), args.Error(1)
}
