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

type ingestService struct {
	kafka *kafkaclient.Client
}

func NewIngestService(kafka *kafkaclient.Client) services.IngestService {
	return &ingestService{kafka: kafka}
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

	// Key by service id so all of a service's records land on the same partition
	// (ordering per service) and load spreads across partitions by service.
	return s.kafka.Writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatUint(service.ID, 10)),
		Value: value,
	})
}
