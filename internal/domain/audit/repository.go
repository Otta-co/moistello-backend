package audit

import "context"

type Repository interface {
	Log(ctx context.Context, entry *AuditEntry) error
}
