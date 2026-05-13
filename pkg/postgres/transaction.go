package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type TxFunc func(tx *sqlx.Tx) error

func WithTransaction(ctx context.Context, db *sqlx.DB, fn TxFunc) error {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func WithTransactionLevel(ctx context.Context, db *sqlx.DB, level sql.IsolationLevel, fn TxFunc) error {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: level})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

type Transactor struct {
	DB *sqlx.DB
}

func NewTransactor(db *sqlx.DB) *Transactor {
	return &Transactor{DB: db}
}

func (t *Transactor) WithTransaction(ctx context.Context, fn func() error) error {
	tx, err := t.DB.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
