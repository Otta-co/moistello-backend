package invite

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Invite struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	CircleID  uuid.UUID    `json:"circleId" db:"circle_id"`
	Code      string       `json:"code" db:"code"`
	CreatedBy uuid.UUID    `json:"createdBy" db:"created_by"`
	MaxUses   int          `json:"maxUses" db:"max_uses"`
	UseCount  int          `json:"useCount" db:"use_count"`
	ExpiresAt sql.NullTime `json:"expiresAt,omitempty" db:"expires_at"`
	CreatedAt time.Time    `json:"createdAt" db:"created_at"`
}
