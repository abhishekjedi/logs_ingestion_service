package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	dbdto "error-logging/db/dto"
	"error-logging/dto"
	kafkaclient "error-logging/pkg/client/kafka"
	"error-logging/services"

	"github.com/segmentio/kafka-go"
)

type messageWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
}

type ingestService struct {
	writer messageWriter
}

func NewIngestService(client *kafkaclient.Client) services.IngestService {
	return &ingestService{writer: client.Writer}
}

func (s *ingestService) Ingest(ctx context.Context, service *dbdto.Service, payload []byte) error {
	msg := dto.LogIngestMessage{
		ServiceID:  service.ID,
		ProjectID:  service.ProjectID,
		ReceivedAt: time.Now().UTC(),
		Payload:    json.RawMessage(payload),
	}

	value, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal ingest message: %w", err)
	}

	return s.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatUint(service.ID, 10)),
		Value: value,
	})
}
