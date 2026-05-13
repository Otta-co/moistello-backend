package contribution

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/moistello/backend/pkg/apperrors"
)

type Service interface {
	Record(ctx context.Context, input RecordInput) (*Contribution, error)
	GetUserHistory(ctx context.Context, userID string, page, limit int) ([]Contribution, int, error)
	GetCircleHistory(ctx context.Context, circleID string, page, limit int) ([]Contribution, int, error)
}

type Transactor interface {
	WithTransaction(ctx context.Context, fn func(repo Repository) error) error
}

type RecordInput struct {
	CircleID    string  `json:"circleId" validate:"required"`
	UserID      string  `json:"userId" validate:"required"`
	RoundNumber int     `json:"roundNumber" validate:"required,gte=1"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	TxnHash     string  `json:"txnHash" validate:"required"`
}

type contributionService struct {
	repo Repository
	tx   Transactor
}

func NewService(repo Repository, tx Transactor) Service {
	return &contributionService{repo: repo, tx: tx}
}

type contribTransactor struct {
	db *sqlx.DB
}

func NewTransactor(db *sqlx.DB) Transactor {
	return &contribTransactor{db: db}
}

func (t *contribTransactor) WithTransaction(ctx context.Context, fn func(repo Repository) error) error {
	tx, err := t.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(NewRepositoryFromTx(tx)); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func parseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID: %w", err)
	}
	return id, nil
}

func (s *contributionService) Record(ctx context.Context, input RecordInput) (*Contribution, error) {
	userID, err := parseUUID(input.UserID)
	if err != nil {
		return nil, err
	}
	circleID, err := parseUUID(input.CircleID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	txnHash := sql.NullString{String: input.TxnHash, Valid: true}

	c := &Contribution{
		ID:          uuid.New(),
		CircleID:    circleID,
		UserID:      userID,
		RoundNumber: input.RoundNumber,
		Amount:      input.Amount,
		TxnHash:     txnHash,
		Status:      StatusPending,
		OnTime:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if s.tx != nil {
		err := s.tx.WithTransaction(ctx, func(repo Repository) error {
			if err := repo.Create(ctx, c); err != nil {
				if err == apperrors.ErrConflict {
					return fmt.Errorf("duplicate contribution: %w", err)
				}
				return fmt.Errorf("recording contribution: %w", err)
			}
			return nil
		})
		return c, err
	}

	if err := s.repo.Create(ctx, c); err != nil {
		if err == apperrors.ErrConflict {
			return nil, fmt.Errorf("duplicate contribution: %w", err)
		}
		return nil, fmt.Errorf("recording contribution: %w", err)
	}
	return c, nil
}

func (s *contributionService) GetUserHistory(ctx context.Context, userID string, page, limit int) ([]Contribution, int, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, 0, err
	}
	contribs, total, err := s.repo.ListByUser(ctx, uid, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("getting user contribution history: %w", err)
	}
	return contribs, total, nil
}

func (s *contributionService) GetCircleHistory(ctx context.Context, circleID string, page, limit int) ([]Contribution, int, error) {
	cid, err := parseUUID(circleID)
	if err != nil {
		return nil, 0, err
	}
	contribs, total, err := s.repo.ListByCircle(ctx, cid, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("getting circle contribution history: %w", err)
	}
	return contribs, total, nil
}
