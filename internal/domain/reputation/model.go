package reputation

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ReputationSnapshot struct {
	UserID    uuid.UUID       `json:"userId" db:"user_id"`
	Score     int             `json:"score" db:"score"`
	Level     string          `json:"level" db:"level"`
	Breakdown json.RawMessage `json:"breakdown" db:"breakdown"`
	Month     time.Time       `json:"month" db:"month"`
	CreatedAt time.Time       `json:"createdAt" db:"created_at"`
}

type ScoreBreakdown struct {
	Streaks     int `json:"streaks"`
	Completions int `json:"completions"`
	Volume      int `json:"volume"`
	Recency     int `json:"recency"`
}
