package services

import "context"

// IssueCache caches fingerprint→issue_id so the hot path skips a MySQL round trip
// once an issue exists. A miss falls through to create/resolve in MySQL.
type IssueCache interface {
	Get(ctx context.Context, serviceID uint64, fingerprint string) (uint64, bool)
	Set(ctx context.Context, serviceID uint64, fingerprint string, issueID uint64)
}
