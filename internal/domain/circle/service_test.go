package circle_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moistello/backend/internal/domain/circle"
	circleMocks "github.com/moistello/backend/internal/domain/circle/mocks"
	"github.com/moistello/backend/pkg/apperrors"
)

func TestCircleService_Create_Success(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()
	orgID := uuid.New().String()

	repo.On("Create", ctx, mock.AnythingOfType("*circle.Circle")).Return(nil)
	repo.On("CreateMember", ctx, mock.AnythingOfType("*circle.CircleMember")).Return(nil)

	c, err := svc.Create(ctx, orgID, circle.CreateCircleInput{
		Name:               "Test Circle",
		CircleType:         circle.CircleTypePublic,
		PayoutType:         circle.PayoutTypeRandom,
		ContributionAmount: 100,
		Currency:           circle.CurrencyUSDC,
		Frequency:          circle.FrequencyWeekly,
		MaxMembers:         10,
		MaxStrikes:         3,
	})

	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "Test Circle", c.Name)
	assert.Equal(t, circle.CircleStatusPending, c.Status)
	assert.Equal(t, 0, c.CurrentRound)
	repo.AssertExpectations(t)
}

func TestCircleService_Create_WithDescription(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()
	orgID := uuid.New().String()

	repo.On("Create", ctx, mock.AnythingOfType("*circle.Circle")).Return(nil)
	repo.On("CreateMember", ctx, mock.AnythingOfType("*circle.CircleMember")).Return(nil)

	c, err := svc.Create(ctx, orgID, circle.CreateCircleInput{
		Name:               "Desc Circle",
		Description:        "A circle with a description",
		CircleType:         circle.CircleTypePublic,
		PayoutType:         circle.PayoutTypeFixed,
		ContributionAmount: 50,
		Currency:           circle.CurrencyUSDC,
		Frequency:          circle.FrequencyMonthly,
		MaxMembers:         5,
		MaxStrikes:         3,
	})

	assert.NoError(t, err)
	assert.True(t, c.Description.Valid)
	assert.Equal(t, "A circle with a description", c.Description.String)
	repo.AssertExpectations(t)
}

func TestCircleService_Create_TooFewMembers(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()
	orgID := uuid.New().String()

	c, err := svc.Create(ctx, orgID, circle.CreateCircleInput{
		Name:               "Small Circle",
		CircleType:         circle.CircleTypePublic,
		PayoutType:         circle.PayoutTypeRandom,
		ContributionAmount: 100,
		Currency:           circle.CurrencyUSDC,
		Frequency:          circle.FrequencyWeekly,
		MaxMembers:         1,
		MaxStrikes:         3,
	})

	assert.Error(t, err)
	assert.Nil(t, c)
	assert.Equal(t, circle.ErrParticipantLimit, err)
}

func TestCircleService_Create_MemberCreationFails(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()
	orgID := uuid.New().String()

	repo.On("Create", ctx, mock.AnythingOfType("*circle.Circle")).Return(nil)
	repo.On("CreateMember", ctx, mock.AnythingOfType("*circle.CircleMember")).Return(assert.AnError)
	repo.On("Delete", ctx, mock.AnythingOfType("uuid.UUID")).Return(nil)

	c, err := svc.Create(ctx, orgID, circle.CreateCircleInput{
		Name:               "Fail Circle",
		CircleType:         circle.CircleTypePublic,
		PayoutType:         circle.PayoutTypeRandom,
		ContributionAmount: 100,
		Currency:           circle.CurrencyUSDC,
		Frequency:          circle.FrequencyWeekly,
		MaxMembers:         10,
		MaxStrikes:         3,
	})

	assert.Error(t, err)
	assert.Nil(t, c)
	repo.AssertExpectations(t)
}

func TestCircleService_Join_Success(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()
	cid := uuid.New().String()
	uid := uuid.New().String()

	c := &circle.Circle{
		ID: uuid.MustParse(cid), Name: "Test", Status: circle.CircleStatusActive,
		MaxMembers: 10, CircleType: circle.CircleTypePublic,
	}
	repo.On("FindByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(c, nil)
	repo.On("GetMemberCount", ctx, mock.AnythingOfType("uuid.UUID")).Return(3, nil)
	repo.On("FindMemberByCircleAndUser", ctx, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(nil, apperrors.ErrNotFound)
	repo.On("CreateMember", ctx, mock.AnythingOfType("*circle.CircleMember")).Return(nil)

	err := svc.Join(ctx, cid, uid, "")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCircleService_Join_Full(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()
	cid := uuid.New().String()
	uid := uuid.New().String()

	c := &circle.Circle{
		ID: uuid.MustParse(cid), Name: "Test", Status: circle.CircleStatusActive,
		MaxMembers: 5, CircleType: circle.CircleTypePublic,
	}
	repo.On("FindByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(c, nil)
	repo.On("GetMemberCount", ctx, mock.AnythingOfType("uuid.UUID")).Return(5, nil)

	err := svc.Join(ctx, cid, uid, "")

	assert.Error(t, err)
	assert.Equal(t, circle.ErrCircleFull, err)
	repo.AssertExpectations(t)
}

func TestCircleService_Join_AlreadyMember(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()
	cid := uuid.New().String()
	uid := uuid.New().String()

	c := &circle.Circle{
		ID: uuid.MustParse(cid), Name: "Test", Status: circle.CircleStatusActive,
		MaxMembers: 10, CircleType: circle.CircleTypePublic,
	}
	existingMember := &circle.CircleMember{
		CircleID: uuid.MustParse(cid), UserID: uuid.MustParse(uid),
		Status: circle.MemberStatusActive,
	}
	repo.On("FindByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(c, nil)
	repo.On("GetMemberCount", ctx, mock.AnythingOfType("uuid.UUID")).Return(3, nil)
	repo.On("FindMemberByCircleAndUser", ctx, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(existingMember, nil)

	err := svc.Join(ctx, cid, uid, "")

	assert.Error(t, err)
	assert.Equal(t, circle.ErrAlreadyMember, err)
	repo.AssertExpectations(t)
}

func TestCircleService_Join_PrivateWithoutInvite(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()
	cid := uuid.New().String()
	uid := uuid.New().String()

	c := &circle.Circle{
		ID: uuid.MustParse(cid), Name: "Private", Status: circle.CircleStatusActive,
		MaxMembers: 10, CircleType: circle.CircleTypePrivate,
	}
	repo.On("FindByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(c, nil)
	repo.On("GetMemberCount", ctx, mock.AnythingOfType("uuid.UUID")).Return(3, nil)
	repo.On("FindMemberByCircleAndUser", ctx, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(nil, apperrors.ErrNotFound)

	err := svc.Join(ctx, cid, uid, "")

	assert.Error(t, err)
	assert.Equal(t, circle.ErrInvalidInvite, err)
	repo.AssertExpectations(t)
}

func TestCircleService_Cancel_Success(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()
	cid := uuid.New()
	orgID := cid
	cidStr := cid.String()
	orgIDStr := orgID.String()

	c := &circle.Circle{
		ID: cid, Name: "Test", Status: circle.CircleStatusPending,
		OrganizerID: orgID,
	}
	repo.On("FindByID", ctx, cid).Return(c, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*circle.Circle")).Return(nil)

	err := svc.Cancel(ctx, cidStr, orgIDStr)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCircleService_Cancel_NotOrganizer(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()
	cid := uuid.New()
	orgID := uuid.New()
	notOrgID := uuid.New().String()

	c := &circle.Circle{
		ID: cid, Name: "Test", Status: circle.CircleStatusPending,
		OrganizerID: orgID,
	}
	repo.On("FindByID", ctx, cid).Return(c, nil)

	err := svc.Cancel(ctx, cid.String(), notOrgID)

	assert.Error(t, err)
	assert.Equal(t, circle.ErrNotOrganizer, err)
	repo.AssertExpectations(t)
}

func TestCircleService_Exit_Success(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()
	cid := uuid.New()
	uid := uuid.New()
	orgID := uuid.New()

	c := &circle.Circle{
		ID: cid, Name: "Test", Status: circle.CircleStatusPending,
		OrganizerID: orgID, TotalContributions: 0,
		ContributionAmount: 100, CollateralPercent: 10,
	}
	member := &circle.CircleMember{
		CircleID: cid, UserID: uid, Status: circle.MemberStatusActive,
	}
	repo.On("FindByID", ctx, cid).Return(c, nil)
	repo.On("FindMemberByCircleAndUser", ctx, cid, uid).Return(member, nil)
	repo.On("UpdateMemberStatus", ctx, cid, uid, circle.MemberStatusExited).Return(nil)

	err := svc.Exit(ctx, cid.String(), uid.String())

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCircleService_Exit_Organizer(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()
	cid := uuid.New()
	orgID := cid

	c := &circle.Circle{
		ID: cid, Name: "Test", Status: circle.CircleStatusPending,
		OrganizerID: orgID,
	}
	repo.On("FindByID", ctx, cid).Return(c, nil)

	err := svc.Exit(ctx, cid.String(), orgID.String())

	assert.Error(t, err)
	assert.Equal(t, circle.ErrNotOrganizer, err)
	repo.AssertExpectations(t)
}

func TestCircleService_GetMembers_Success(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()
	cid := uuid.New()

	members := []circle.CircleMember{
		{CircleID: cid, UserID: uuid.New(), Position: 1, Status: circle.MemberStatusActive, JoinedAt: time.Now()},
		{CircleID: cid, UserID: uuid.New(), Position: 2, Status: circle.MemberStatusActive, JoinedAt: time.Now()},
	}
	repo.On("FindByID", ctx, cid).Return(&circle.Circle{ID: cid, Name: "Test"}, nil)
	repo.On("GetMembers", ctx, cid).Return(members, nil)

	result, err := svc.GetMembers(ctx, cid.String())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	repo.AssertExpectations(t)
}

func TestCircleService_List_Success(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()

	now := time.Now()
	circles := []circle.Circle{
		{ID: uuid.New(), Name: "C1", Status: circle.CircleStatusActive, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), Name: "C2", Status: circle.CircleStatusPending, CreatedAt: now, UpdatedAt: now},
	}
	filter := circle.CircleFilter{Page: 1, Limit: 20}
	repo.On("List", ctx, filter).Return(circles, nil)
	repo.On("Count", ctx, filter).Return(2, nil)

	result, total, err := svc.List(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, total)
	repo.AssertExpectations(t)
}

func TestCircleService_Get_Success(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()
	cid := uuid.New()

	c := &circle.Circle{
		ID: cid, Name: "Test Circle", Status: circle.CircleStatusActive,
		Description: sql.NullString{String: "Desc", Valid: true},
	}
	repo.On("FindByID", ctx, cid).Return(c, nil)

	result, err := svc.Get(ctx, cid.String())

	assert.NoError(t, err)
	assert.Equal(t, "Test Circle", result.Name)
	repo.AssertExpectations(t)
}

func TestCircleService_Get_NotFound(t *testing.T) {
	repo := new(circleMocks.Repository)
	svc := circle.NewService(repo, nil)
	ctx := context.Background()
	cid := uuid.New()

	repo.On("FindByID", ctx, cid).Return(nil, circle.ErrCircleNotFound)

	_, err := svc.Get(ctx, cid.String())

	assert.Error(t, err)
	repo.AssertExpectations(t)
}
