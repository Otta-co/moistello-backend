package audit

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type pgRepo struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &pgRepo{db: db}
}

func (r *pgRepo) Log(ctx context.Context, entry *AuditEntry) error {
	query := `INSERT INTO audit_log (id, actor_id, action, resource_type, resource_id, details, ip_address, user_agent, created_at)
		VALUES (:id, :actor_id, :action, :resource_type, :resource_id, :details, :ip_address, :user_agent, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, entry)
	if err != nil {
		return fmt.Errorf("logging audit entry: %w", err)
	}
	return nil
}
