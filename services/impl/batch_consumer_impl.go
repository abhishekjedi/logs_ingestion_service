package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"error-logging/db/repository"
	"error-logging/dto"
	kafkaclient "error-logging/pkg/client/kafka"
	"error-logging/pkg/config"
	"error-logging/services"

	"github.com/segmentio/kafka-go"
)

const cycleFlushTimeout = 30 * time.Second

type messageReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
}

type batchConsumer struct {
	reader      messageReader
	processor   services.ProcessorService
	logs        repository.LogRepository
	errorEvents repository.ErrorEventRepository
	cfg         config.WorkerConfig
}

func NewBatchConsumer(
	client *kafkaclient.Client,
	processor services.ProcessorService,
	logs repository.LogRepository,
	errorEvents repository.ErrorEventRepository,
	cfg config.WorkerConfig,
) services.BatchConsumer {
	return &batchConsumer{
		reader:      client.Reader,
		processor:   processor,
		logs:        logs,
		errorEvents: errorEvents,
		cfg:         cfg,
	}
}

func (c *batchConsumer) Run(ctx context.Context) error {
	log.Println("batch consumer started")
	for ctx.Err() == nil {
		msgs, kmsgs, shutdown := c.fetchCycle(ctx)
		if len(kmsgs) > 0 {
			if err := c.processCycle(msgs, kmsgs); err != nil {
				return err
			}
		}
		if shutdown {
			break
		}
	}
	log.Println("batch consumer stopped cleanly")
	return nil
}

// fetchCycle builds a bounded micro-batch: one blocking fetch, then a short
// accumulation window capped by message count and bytes.
func (c *batchConsumer) fetchCycle(ctx context.Context) (msgs []dto.LogIngestMessage, kmsgs []kafka.Message, shutdown bool) {
	first, err := c.reader.FetchMessage(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return nil, nil, true
		}
		log.Printf("fetch error: %v", err)
		return nil, nil, false
	}
	accumulate(first, &msgs, &kmsgs)
	bytes := len(first.Value)

	deadline := time.Now().Add(c.cfg.FetchMaxWait)
	for len(kmsgs) < c.cfg.FetchMaxMessages && bytes < c.cfg.FetchMaxBytes {
		fctx, cancel := context.WithDeadline(ctx, deadline)
		m, err := c.reader.FetchMessage(fctx)
		cancel()
		if err != nil {
			break
		}
		accumulate(m, &msgs, &kmsgs)
		bytes += len(m.Value)
	}
	return msgs, kmsgs, false
}

func accumulate(km kafka.Message, msgs *[]dto.LogIngestMessage, kmsgs *[]kafka.Message) {
	*kmsgs = append(*kmsgs, km)
	var m dto.LogIngestMessage
	if err := json.Unmarshal(km.Value, &m); err != nil {
		log.Printf("skipping malformed message: %v", err)
		return
	}
	*msgs = append(*msgs, m)
}

func (c *batchConsumer) processCycle(msgs []dto.LogIngestMessage, kmsgs []kafka.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), cycleFlushTimeout)
	defer cancel()

	var result dto.TransformResult
	if len(msgs) > 0 {
		r, err := c.processor.TransformBatch(ctx, msgs)
		if err != nil {
			return fmt.Errorf("transform cycle: %w", err)
		}
		result = r
	}

	if err := c.flush(ctx, result); err != nil {
		return fmt.Errorf("flush cycle: %w", err)
	}

	if err := c.reader.CommitMessages(ctx, kmsgs...); err != nil {
		return fmt.Errorf("commit cycle: %w", err)
	}

	log.Printf("cycle committed: %d messages, %d logs, %d error_events",
		len(kmsgs), len(result.Logs), len(result.ErrorEvents))
	return nil
}

// flush chunks ClickHouse writes so the driver's batch buffer never holds a whole
// cycle; a failed chunk aborts before Kafka commit, so replay stays at-least-once.
func (c *batchConsumer) flush(ctx context.Context, res dto.TransformResult) error {
	if err := c.flushChunked(len(res.Logs), func(lo, hi int) error {
		return c.logs.InsertBatch(ctx, res.Logs[lo:hi])
	}); err != nil {
		return fmt.Errorf("logs: %w", err)
	}
	if err := c.flushChunked(len(res.ErrorEvents), func(lo, hi int) error {
		return c.errorEvents.InsertBatch(ctx, res.ErrorEvents[lo:hi])
	}); err != nil {
		return fmt.Errorf("error_events: %w", err)
	}
	return nil
}

func (c *batchConsumer) flushChunked(total int, insert func(lo, hi int) error) error {
	if total == 0 {
		return nil
	}
	chunk := c.cfg.FlushChunkRows
	if chunk <= 0 {
		chunk = total
	}
	for lo := 0; lo < total; lo += chunk {
		hi := lo + chunk
		if hi > total {
			hi = total
		}
		if err := c.retryInsert(func() error { return insert(lo, hi) }); err != nil {
			return err
		}
	}
	return nil
}

func (c *batchConsumer) retryInsert(insert func() error) error {
	var lastErr error
	for attempt := 1; attempt <= c.cfg.FlushRetries; attempt++ {
		if err := insert(); err != nil {
			lastErr = err
			log.Printf("flush attempt %d/%d failed, retrying: %v", attempt, c.cfg.FlushRetries, err)
			time.Sleep(time.Duration(attempt) * 250 * time.Millisecond)
			continue
		}
		return nil
	}
	return lastErr
}

var _ messageReader = (*kafka.Reader)(nil)
