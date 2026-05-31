package processors

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"inventory-transaction-service/config"
	"inventory-transaction-service/models"
	"inventory-transaction-service/observability"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

// InventoryProducer mirrors InventoryProcessor from inventory-service.
// Uses SASL/SCRAM-SHA512 over TLS (SASL_SSL), verified with ca.pem.
type InventoryProducer struct {
	writer     *kafka.Writer
	dialer     *kafka.Dialer
	serializer *AvroSerializer
	metrics    *observability.PrometheusMetrics
}

// NewInventoryProducer creates a bare InventoryProducer. Call Init before use.
func NewInventoryProducer(metrics *observability.PrometheusMetrics) *InventoryProducer {
	return &InventoryProducer{metrics: metrics}
}

// Init mirrors InventoryProcessor.Init from inventory-service exactly:
//  1. Initialize Avro serializer
//  2. Read CA certificate (KAFKA_CA_FILE_PATH)
//  3. Create TLS config with RootCAs
//  4. Create SCRAM-SHA512 mechanism (KAFKA_USERNAME / KAFKA_PASSWORD)
//  5. Build kafka.Dialer (same fields as inventory-service)
//  6. Init writer — producer equivalent of kafka.NewReader in inventory-service
func (p *InventoryProducer) Init(cfg config.Config) error {
	var registry *SchemaRegistryClient
	if cfg.SchemaRegistryURL != "" {
		registry = NewSchemaRegistryClient(
			cfg.SchemaRegistryURL,
			cfg.SchemaRegistryUsername,
			cfg.SchemaRegistryPassword,
		)
	}
	p.serializer = NewAvroSerializer(registry, cfg.SchemaSubject, cfg.KafkaAutoRegister)

	caCert, err := os.ReadFile(cfg.KafkaCAFilePath)
	if err != nil {
		log.Fatalf("Failed to read CA certificate file: %s", err)
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		log.Fatalf("Failed to parse CA certificate file: %s", err)
	}

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	scramMechanism, err := scram.Mechanism(scram.SHA512, cfg.KafkaUsername, cfg.KafkaPassword)
	if err != nil {
		log.Fatalf("Failed to create scram mechanism: %s", err)
	}

	p.dialer = &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		TLS:           tlsConfig,
		SASLMechanism: scramMechanism,
	}

	p.writer = &kafka.Writer{
		Addr:                   kafka.TCP(splitBrokers(cfg.KafkaBootstrapServer)...),
		Topic:                  cfg.KafkaTopicName,
		Balancer:               &kafka.Hash{},
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: false,
		WriteTimeout:           10 * time.Second,
		Transport: &kafka.Transport{
			TLS:         tlsConfig,
			SASL:        scramMechanism,
			DialTimeout: 10 * time.Second,
		},
	}

	slog.Info("Kafka producer initialised",
		slog.String("topic", cfg.KafkaTopicName),
		slog.String("brokers", cfg.KafkaBootstrapServer))

	return nil
}

// Send serializes the event to Avro and publishes it keyed by `key`.
func (p *InventoryProducer) Send(ctx context.Context, key string, event models.InventoryImportEvent) error {
	payload, err := p.serializer.Serialize(event)
	if err != nil {
		p.record("encode_error")
		return fmt.Errorf("serialize event: %w", err)
	}

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now(),
	}); err != nil {
		p.record("error")
		return fmt.Errorf("write to kafka: %w", err)
	}

	slog.Info("Inventory status update published",
		slog.String("topic", p.writer.Stats().Topic),
		slog.String("inventoryId", key))
	p.record("success")
	return nil
}

// Close releases the underlying Kafka writer.
func (p *InventoryProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// CreateInventoryImportEvent assembles an InventoryImportEvent from a saved record.
func CreateInventoryImportEvent(record models.TransactionRecord) models.InventoryImportEvent {
	return models.InventoryImportEvent{
		InventoryID:       record.InventoryId,
		InventoryMeasure:  record.InventoryMeasure,
		InventoryCategory: record.InventoryCategory,
		InventoryUnit:     record.InventoryUnit,
		Quantity:          int(record.Quantity),
		Type:              record.Type,
	}
}

func (p *InventoryProducer) record(status string) {
	if p.metrics != nil && p.writer != nil {
		p.metrics.RecordKafkaPublish(p.writer.Stats().Topic, status)
	}
}

func splitBrokers(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, b := range parts {
		if trimmed := strings.TrimSpace(b); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
