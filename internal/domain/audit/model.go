package audit

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditEntry struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	ActorID      uuid.UUID       `json:"actorId" db:"actor_id"`
	Action       string          `json:"action" db:"action"`
	ResourceType string          `json:"resourceType" db:"resource_type"`
	ResourceID   sql.NullString  `json:"resourceId,omitempty" db:"resource_id"`
	Details      json.RawMessage `json:"details,omitempty" db:"details"`
	IPAddress    sql.NullString  `json:"ipAddress,omitempty" db:"ip_address"`
	UserAgent    sql.NullString  `json:"userAgent,omitempty" db:"user_agent"`
	CreatedAt    time.Time       `json:"createdAt" db:"created_at"`
}
