package contribution

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ContributionStatus string

const (
	StatusPending   ContributionStatus = "pending"
	StatusConfirmed ContributionStatus = "confirmed"
	StatusFailed    ContributionStatus = "failed"
	StatusLate      ContributionStatus = "late"
)

type Contribution struct {
	ID          uuid.UUID          `json:"id" db:"id"`
	CircleID    uuid.UUID          `json:"circleId" db:"circle_id"`
	UserID      uuid.UUID          `json:"userId" db:"user_id"`
	RoundNumber int                `json:"roundNumber" db:"round_number"`
	Amount      float64            `json:"amount" db:"amount"`
	TxnHash     sql.NullString     `json:"txnHash,omitempty" db:"txn_hash"`
	Status      ContributionStatus `json:"status" db:"status"`
	OnTime      bool               `json:"onTime" db:"on_time"`
	CreatedAt   time.Time          `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time          `json:"updatedAt" db:"updated_at"`
}
