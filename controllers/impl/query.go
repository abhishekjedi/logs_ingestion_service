package impl

import (
	"strconv"
	"time"

	"error-logging/pkg/context"
)

func parseTimeRange(c *context.ApiContext) (from, to time.Time) {
	to = time.Now().UTC()
	from = to.Add(-24 * time.Hour)
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = t
		}
	}
	return from, to
}

func queryInt(c *context.ApiContext, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func uintParam(c *context.ApiContext, key string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(key), 10, 64)
	return id, err == nil
}
