package reputation_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moistello/backend/internal/domain/reputation"
	repMocks "github.com/moistello/backend/internal/domain/reputation/mocks"
)

func TestReputationService_CalculateScore_Bronze(t *testing.T) {
	repo := new(repMocks.Repository)
	svc := reputation.NewService(repo)
	ctx := context.Background()
	userID := uuid.New().String()

	snapshot, err := svc.CalculateScore(ctx, userID, 1, 1, 10, 30)

	assert.NoError(t, err)
	assert.NotNil(t, snapshot)
	assert.Equal(t, "Bronze", snapshot.Level)
	assert.GreaterOrEqual(t, snapshot.Score, 0)
	assert.LessOrEqual(t, snapshot.Score, 1000)

	var breakdown reputation.ScoreBreakdown
	err = json.Unmarshal(snapshot.Breakdown, &breakdown)
	assert.NoError(t, err)
}

func TestReputationService_CalculateScore_Silver(t *testing.T) {
	repo := new(repMocks.Repository)
	svc := reputation.NewService(repo)
	ctx := context.Background()
	userID := uuid.New().String()

	snapshot, err := svc.CalculateScore(ctx, userID, 3, 2, 100, 20)

	assert.NoError(t, err)
	assert.Equal(t, "Silver", snapshot.Level)
	assert.Greater(t, snapshot.Score, 200)
	assert.LessOrEqual(t, snapshot.Score, 400)
}

func TestReputationService_CalculateScore_Gold(t *testing.T) {
	repo := new(repMocks.Repository)
	svc := reputation.NewService(repo)
	ctx := context.Background()
	userID := uuid.New().String()

	snapshot, err := svc.CalculateScore(ctx, userID, 4, 3, 300, 15)

	assert.NoError(t, err)
	assert.Equal(t, "Gold", snapshot.Level)
	assert.Greater(t, snapshot.Score, 400)
	assert.LessOrEqual(t, snapshot.Score, 600)
}

func TestReputationService_CalculateScore_Platinum(t *testing.T) {
	repo := new(repMocks.Repository)
	svc := reputation.NewService(repo)
	ctx := context.Background()
	userID := uuid.New().String()

	snapshot, err := svc.CalculateScore(ctx, userID, 6, 5, 1000, 5)

	assert.NoError(t, err)
	assert.Equal(t, "Platinum", snapshot.Level)
	assert.Greater(t, snapshot.Score, 600)
	assert.LessOrEqual(t, snapshot.Score, 800)
}

func TestReputationService_CalculateScore_Diamond(t *testing.T) {
	repo := new(repMocks.Repository)
	svc := reputation.NewService(repo)
	ctx := context.Background()
	userID := uuid.New().String()

	snapshot, err := svc.CalculateScore(ctx, userID, 8, 6, 5000, 0)

	assert.NoError(t, err)
	assert.Equal(t, "Diamond", snapshot.Level)
	assert.Greater(t, snapshot.Score, 800)
}

func TestReputationService_CalculateScore_CappedAt1000(t *testing.T) {
	repo := new(repMocks.Repository)
	svc := reputation.NewService(repo)
	ctx := context.Background()
	userID := uuid.New().String()

	snapshot, err := svc.CalculateScore(ctx, userID, 50, 50, 1000000, 0)

	assert.NoError(t, err)
	assert.LessOrEqual(t, snapshot.Score, 1000)
	assert.Equal(t, "Diamond", snapshot.Level)
}

func TestReputationService_CalculateScore_InvalidUUID(t *testing.T) {
	repo := new(repMocks.Repository)
	svc := reputation.NewService(repo)
	ctx := context.Background()

	snapshot, err := svc.CalculateScore(ctx, "not-a-uuid", 1, 1, 10, 30)

	assert.Error(t, err)
	assert.Nil(t, snapshot)
}

func TestReputationService_UpdateScore_Success(t *testing.T) {
	repo := new(repMocks.Repository)
	svc := reputation.NewService(repo)
	ctx := context.Background()
	userID := uuid.New().String()

	repo.On("SaveSnapshot", ctx, mock.AnythingOfType("*reputation.ReputationSnapshot")).Return(nil)

	snapshot, err := svc.UpdateScore(ctx, userID, 3, 2, 100, 20)

	assert.NoError(t, err)
	assert.NotNil(t, snapshot)
	assert.Equal(t, "Silver", snapshot.Level)
	repo.AssertExpectations(t)
}

func TestReputationService_UpdateScore_SaveFails(t *testing.T) {
	repo := new(repMocks.Repository)
	svc := reputation.NewService(repo)
	ctx := context.Background()
	userID := uuid.New().String()

	repo.On("SaveSnapshot", ctx, mock.AnythingOfType("*reputation.ReputationSnapshot")).Return(assert.AnError)

	snapshot, err := svc.UpdateScore(ctx, userID, 3, 2, 100, 20)

	assert.Error(t, err)
	assert.Nil(t, snapshot)
	repo.AssertExpectations(t)
}
