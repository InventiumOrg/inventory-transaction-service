package processors

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"

	"inventory-transaction-service/models"

	"github.com/linkedin/goavro/v2"
)

// AvroSerializer serializes InventoryImportEvent records using a Confluent
// Schema-Registry compatible wire format:
//
//	[ magic byte (0x00) | schema id (uint32 BE) | avro binary payload ]
type AvroSerializer struct {
	registry      *SchemaRegistryClient
	subject       string
	autoRegister  bool
	fallbackAvro  string
	mu            sync.RWMutex
	cachedID      int
	cachedCodec   *goavro.Codec
	schemaResolved bool
}

// FallbackInventoryImportSchema is used when the registry cannot be reached.
// It matches the Java producer's `inventory.transaction.status.updated` shape.
const FallbackInventoryImportSchema = `{
  "type": "record",
  "name": "inventory_import",
  "namespace": "com.inventium",
  "fields": [
    {"name": "quantity",          "type": "int",    "default": 0},
    {"name": "inventoryId",       "type": "string", "default": ""},
    {"name": "inventoryMeasure",  "type": "string", "default": ""},
    {"name": "inventoryCategory", "type": "string", "default": ""},
    {"name": "inventoryUnit",     "type": "string", "default": ""},
    {"name": "type",              "type": "string", "default": ""}
  ]
}`

// NewAvroSerializer wires the serializer to a Schema Registry client and the
// target subject. If `autoRegister` is true the fallback schema is registered
// on demand when no version exists for the subject.
func NewAvroSerializer(registry *SchemaRegistryClient, subject string, autoRegister bool) *AvroSerializer {
	return &AvroSerializer{
		registry:     registry,
		subject:      subject,
		autoRegister: autoRegister,
		fallbackAvro: FallbackInventoryImportSchema,
	}
}

// Serialize encodes the InventoryImportEvent using the resolved Avro schema.
func (s *AvroSerializer) Serialize(event models.InventoryImportEvent) ([]byte, error) {
	id, codec, err := s.resolveSchema()
	if err != nil {
		return nil, err
	}

	native := map[string]interface{}{
		"quantity":          int32(event.Quantity),
		"inventoryId":       event.InventoryID,
		"inventoryMeasure":  event.InventoryMeasure,
		"inventoryCategory": event.InventoryCategory,
		"inventoryUnit":     event.InventoryUnit,
		"type":              event.Type,
	}

	avroBinary, err := codec.BinaryFromNative(nil, native)
	if err != nil {
		return nil, fmt.Errorf("avro encode: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteByte(0)
	if err := binary.Write(&buf, binary.BigEndian, uint32(id)); err != nil {
		return nil, fmt.Errorf("write schema id: %w", err)
	}
	buf.Write(avroBinary)
	return buf.Bytes(), nil
}

func (s *AvroSerializer) resolveSchema() (int, *goavro.Codec, error) {
	s.mu.RLock()
	if s.schemaResolved {
		id, codec := s.cachedID, s.cachedCodec
		s.mu.RUnlock()
		return id, codec, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.schemaResolved {
		return s.cachedID, s.cachedCodec, nil
	}

	id, schemaJSON, err := s.fetchOrRegister()
	if err != nil {
		return 0, nil, err
	}
	codec, err := goavro.NewCodec(schemaJSON)
	if err != nil {
		return 0, nil, fmt.Errorf("build avro codec: %w", err)
	}
	s.cachedID = id
	s.cachedCodec = codec
	s.schemaResolved = true
	return id, codec, nil
}

func (s *AvroSerializer) fetchOrRegister() (int, string, error) {
	if s.registry != nil {
		if resp, err := s.registry.GetLatestSchema(s.subject); err == nil {
			return resp.ID, resp.Schema, nil
		}
		if s.autoRegister {
			id, err := s.registry.RegisterSchema(s.subject, s.fallbackAvro)
			if err == nil {
				return id, s.fallbackAvro, nil
			}
		}
	}
	// As a last resort, use the fallback schema with id=0 so encoding still
	// works in environments without a reachable registry (e.g. local dev).
	return 0, s.fallbackAvro, nil
}
