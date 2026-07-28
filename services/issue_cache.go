package services

import "context"

type IssueCache interface {
	Get(ctx context.Context, serviceID uint64, fingerprint string) (uint64, bool)
	Set(ctx context.Context, serviceID uint64, fingerprint string, issueID uint64)
}
