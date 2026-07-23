package impl

import (
	"context"
	"fmt"

	"error-logging/db/repository"
	chclient "error-logging/pkg/client/clickhouse"
)

type errorEventRepository struct {
	conn *chclient.NativeClient
}

func NewErrorEventRepository(c *chclient.NativeClient) repository.ErrorEventRepository {
	return &errorEventRepository{conn: c}
}

// InsertBatch writes full-fidelity error rows. stack_frames is passed as positional
// tuples ([]any per frame) matching Array(Tuple(file, function, line, col, in_app)).
func (r *errorEventRepository) InsertBatch(ctx context.Context, rows []repository.ErrorEventRow) error {
	if len(rows) == 0 {
		return nil
	}

	batch, err := r.conn.Conn.PrepareBatch(ctx, "INSERT INTO error_events")
	if err != nil {
		return fmt.Errorf("prepare error_events batch: %w", err)
	}

	for _, row := range rows {
		frames := make([][]any, len(row.StackFrames))
		for i, f := range row.StackFrames {
			frames[i] = []any{f.File, f.Function, f.Line, f.Col, f.InApp}
		}

		if err := batch.Append(
			row.EventID,
			row.IssueID,
			row.ServiceID,
			row.ProjectID,
			row.Timestamp,
			row.IngestedAt,
			row.SeverityNumber,
			row.SeverityText,
			row.Environment,
			row.Release,
			row.ExceptionType,
			row.ExceptionMessage,
			row.UserID,
			row.SessionID,
			frames,
			row.RawStacktrace,
			row.TraceID,
			row.SpanID,
			row.Attributes,
			row.ResourceAttributes,
			row.S3Key,
		); err != nil {
			return fmt.Errorf("append error_events row: %w", err)
		}
	}

	return batch.Send()
}
