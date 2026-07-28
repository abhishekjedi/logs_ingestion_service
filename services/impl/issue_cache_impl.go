package impl

import (
	"context"
	"fmt"
	"time"

	redisclient "error-logging/pkg/client/redis"
	"error-logging/services"
)

const issueCacheTTL = 5 * time.Minute

type issueCache struct {
	redis *redisclient.Client
}

func NewIssueCache(redis *redisclient.Client) services.IssueCache {
	return &issueCache{redis: redis}
}

func (c *issueCache) Get(ctx context.Context, serviceID uint64, fingerprint string) (uint64, bool) {
	if c.redis == nil || c.redis.RDB == nil {
		return 0, false
	}
	id, err := c.redis.RDB.Get(ctx, c.key(serviceID, fingerprint)).Uint64()
	if err != nil {
		return 0, false
	}
	return id, true
}

func (c *issueCache) Set(ctx context.Context, serviceID uint64, fingerprint string, issueID uint64) {
	if c.redis == nil || c.redis.RDB == nil {
		return
	}
	c.redis.RDB.Set(ctx, c.key(serviceID, fingerprint), issueID, issueCacheTTL)
}

func (c *issueCache) key(serviceID uint64, fingerprint string) string {
	return fmt.Sprintf("issue:%d:%s", serviceID, fingerprint)
}
