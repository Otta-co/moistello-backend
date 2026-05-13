package invite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/moistello/backend/pkg/apperrors"
)

type pgRepo struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &pgRepo{db: db}
}

func scanInvite(row interface{ Scan(...interface{}) error }) (*Invite, error) {
	var i Invite
	var expiresAt sql.NullTime
	err := row.Scan(&i.ID, &i.CircleID, &i.Code, &i.CreatedBy, &i.MaxUses, &i.UseCount, &expiresAt, &i.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("scanning invite row: %w", err)
	}
	i.ExpiresAt = expiresAt
	return &i, nil
}

func (r *pgRepo) FindByCode(ctx context.Context, code string) (*Invite, error) {
	query := `SELECT id, circle_id, code, created_by, max_uses, use_count, expires_at, created_at
		FROM invites WHERE code = $1`
	return scanInvite(r.db.QueryRowxContext(ctx, query, code))
}

func (r *pgRepo) FindByCircle(ctx context.Context, circleID uuid.UUID) ([]Invite, error) {
	query := `SELECT id, circle_id, code, created_by, max_uses, use_count, expires_at, created_at
		FROM invites WHERE circle_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryxContext(ctx, query, circleID)
	if err != nil {
		return nil, fmt.Errorf("finding circle invites: %w", err)
	}
	defer rows.Close()

	var invites []Invite
	for rows.Next() {
		i, err := scanInvite(rows)
		if err != nil {
			return nil, err
		}
		invites = append(invites, *i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating invites: %w", err)
	}
	return invites, nil
}

func (r *pgRepo) Create(ctx context.Context, i *Invite) error {
	query := `INSERT INTO invites (id, circle_id, code, created_by, max_uses, use_count, expires_at, created_at)
		VALUES (:id, :circle_id, :code, :created_by, :max_uses, :use_count, :expires_at, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, i)
	if err != nil {
		return fmt.Errorf("creating invite: %w", err)
	}
	return nil
}

func (r *pgRepo) IncrementUse(ctx context.Context, code string) error {
	query := `UPDATE invites SET use_count = use_count + 1 WHERE code = $1`
	result, err := r.db.ExecContext(ctx, query, code)
	if err != nil {
		return fmt.Errorf("incrementing invite use: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *pgRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM invites WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting invite: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
