package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// Cursor tracks the last processed ledger for a given chain.
type Cursor struct {
	Chain           string    `db:"chain" json:"chain"`
	LastLedger      int64     `db:"last_ledger" json:"lastLedger"`
	LastProcessedAt time.Time `db:"last_processed_at" json:"lastProcessedAt"`
}

// CursorTracker persists and retrieves the indexer cursor in PostgreSQL.
type CursorTracker struct {
	db *sqlx.DB
}

// NewCursorTracker creates a new CursorTracker backed by the given database.
func NewCursorTracker(db *sqlx.DB) *CursorTracker {
	return &CursorTracker{db: db}
}

// GetCurrent reads the current cursor from the database.
func (c *CursorTracker) GetCurrent(ctx context.Context) (*Cursor, error) {
	var cursor Cursor
	err := c.db.GetContext(ctx, &cursor, "SELECT chain, last_ledger, last_processed_at FROM indexer_cursor WHERE chain = 'stellar'")
	if err != nil {
		return nil, fmt.Errorf("reading cursor: %w", err)
	}
	return &cursor, nil
}

// Update writes the new cursor position after successful processing.
func (c *CursorTracker) Update(ctx context.Context, lastLedger int64) error {
	_, err := c.db.ExecContext(ctx,
		"UPDATE indexer_cursor SET last_ledger = $1, last_processed_at = $2 WHERE chain = 'stellar'",
		lastLedger, time.Now())
	if err != nil {
		return fmt.Errorf("updating cursor: %w", err)
	}
	return nil
}
