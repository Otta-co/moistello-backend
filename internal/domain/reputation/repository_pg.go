package reputation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/moistello/backend/pkg/apperrors"
)

type pgRepo struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &pgRepo{db: db}
}

func scanSnapshot(row interface{ Scan(...interface{}) error }) (*ReputationSnapshot, error) {
	var s ReputationSnapshot
	var breakdown json.RawMessage
	err := row.Scan(&s.UserID, &s.Score, &s.Level, &breakdown, &s.Month, &s.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("scanning reputation snapshot: %w", err)
	}
	s.Breakdown = breakdown
	return &s, nil
}

func (r *pgRepo) GetByUser(ctx context.Context, userID uuid.UUID) (*ReputationSnapshot, error) {
	query := `SELECT user_id, score, level, breakdown, month, created_at
		FROM reputation_snapshots WHERE user_id = $1 ORDER BY month DESC LIMIT 1`
	return scanSnapshot(r.db.QueryRowxContext(ctx, query, userID))
}

func (r *pgRepo) SaveSnapshot(ctx context.Context, s *ReputationSnapshot) error {
	query := `INSERT INTO reputation_snapshots (user_id, score, level, breakdown, month, created_at)
		VALUES (:user_id, :score, :level, :breakdown, :month, :created_at)
		ON CONFLICT (user_id, month) DO UPDATE SET score = :score, level = :level, breakdown = :breakdown, created_at = :created_at`
	_, err := r.db.NamedExecContext(ctx, query, s)
	if err != nil {
		if isUniqueViolationPg(err) {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("saving reputation snapshot: %w", err)
	}
	return nil
}

func (r *pgRepo) GetHistory(ctx context.Context, userID uuid.UUID, months int) ([]ReputationSnapshot, error) {
	if months < 1 {
		months = 12
	}
	since := time.Now().UTC().AddDate(0, -months, 0)
	query := `SELECT user_id, score, level, breakdown, month, created_at
		FROM reputation_snapshots WHERE user_id = $1 AND month >= $2 ORDER BY month ASC`
	rows, err := r.db.QueryxContext(ctx, query, userID, since)
	if err != nil {
		return nil, fmt.Errorf("getting reputation history: %w", err)
	}
	defer rows.Close()

	var snapshots []ReputationSnapshot
	for rows.Next() {
		s, err := scanSnapshot(rows)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, *s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating snapshots: %w", err)
	}
	return snapshots, nil
}

func isUniqueViolationPg(err error) bool {
	if err == nil {
		return false
	}
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == pq.ErrorCode("23505")
	}
	return false
}
