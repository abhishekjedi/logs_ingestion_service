package impl

import (
	"context"
	"fmt"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	chclient "error-logging/pkg/client/clickhouse"
)

type logRepository struct {
	conn *chclient.NativeClient
}

func NewLogRepository(c *chclient.NativeClient) repository.LogRepository {
	return &logRepository{conn: c}
}

// InsertBatch appends all rows into one native batch. Column order matches the
// logs table DDL. Writes go to the Distributed `logs` table, which routes to the
// local shard where the materialized views fire.
func (r *logRepository) InsertBatch(ctx context.Context, rows []dbdto.LogRow) error {
	if len(rows) == 0 {
		return nil
	}

	batch, err := r.conn.Conn.PrepareBatch(ctx, "INSERT INTO logs")
	if err != nil {
		return fmt.Errorf("prepare logs batch: %w", err)
	}

	for _, row := range rows {
		if err := batch.Append(
			row.Timestamp,
			row.ObservedAt,
			row.ProjectID,
			row.ServiceID,
			row.IssueID,
			row.SeverityNumber,
			row.SeverityText,
			row.Body,
			row.TraceID,
			row.SpanID,
			row.Environment,
			row.Release,
			row.UserID,
			row.SessionID,
			row.ExceptionType,
			row.ExceptionMessage,
			row.Attributes,
			row.ResourceAttributes,
		); err != nil {
			return fmt.Errorf("append logs row: %w", err)
		}
	}

	return batch.Send()
}
