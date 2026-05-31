package config

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all runtime configuration for the inventory-transaction-service.
// Values are loaded from environment variables (and optionally an `app.env` file
// when present in the working directory).
type Config struct {
	ServiceName string `mapstructure:"SERVICE_NAME"`
	ServerPort  string `mapstructure:"SERVER_PORT"`

	// AWS / DynamoDB
	AWSRegion         string `mapstructure:"AWS_REGION"`
	DynamoDBEndpoint  string `mapstructure:"DYNAMODB_ENDPOINT"`
	DynamoDBTableName string `mapstructure:"DYNAMODB_TABLE_NAME"`

	// Kafka (inventory status updates)
	KafkaTopicName       string `mapstructure:"KAFKA_TOPIC_NAME"`
	KafkaBootstrapServer string `mapstructure:"KAFKA_BOOTSTRAP_SERVERS"`
	KafkaCAFilePath      string `mapstructure:"KAFKA_CA_FILE_PATH"`
	KafkaSvcCertLocation string `mapstructure:"KAFKA_SVC_CERT_LOCATION"`
	KafkaSvcKeyLocation  string `mapstructure:"KAFKA_SVC_KEY_LOCATION"`
	KafkaAutoRegister    bool   `mapstructure:"KAFKA_AUTO_REGISTER_SCHEMAS"`

	// Schema Registry
	SchemaRegistryURL      string `mapstructure:"SCHEMA_REGISTRY_URL"`
	SchemaRegistryUsername string `mapstructure:"SCHEMA_REGISTRY_USERNAME"`
	SchemaRegistryPassword string `mapstructure:"SCHEMA_REGISTRY_PASSWORD"`
	SchemaSubject          string `mapstructure:"KAFKA_INVENTORY_SCHEMA_SUBJECT"`

	// Observability
	OTELExporterOTLPEndpoint string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTELExporterOTLPHeaders  string `mapstructure:"OTEL_EXPORTER_OTLP_HEADERS"`
	OTELResourceAttributes   string `mapstructure:"OTEL_RESOURCE_ATTRIBUTES"`
	TracingSamplingRatio     float64 `mapstructure:"TRACING_SAMPLING_PROBABILITY"`
}

// LoadConfig loads configuration from environment variables.
// If `app.env` exists in `path`, values from it are loaded first and then
// overridden by environment variables.
func LoadConfig(path string) (Config, error) {
	var cfg Config

	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.AutomaticEnv()

	for _, key := range []string{
		"SERVICE_NAME", "SERVER_PORT",
		"AWS_REGION", "DYNAMODB_ENDPOINT", "DYNAMODB_TABLE_NAME",
		"KAFKA_TOPIC_NAME", "KAFKA_BOOTSTRAP_SERVERS",
		"KAFKA_CA_FILE_PATH", "KAFKA_SVC_CERT_LOCATION", "KAFKA_SVC_KEY_LOCATION", "KAFKA_AUTO_REGISTER_SCHEMAS",
		"SCHEMA_REGISTRY_URL", "SCHEMA_REGISTRY_USERNAME", "SCHEMA_REGISTRY_PASSWORD",
		"KAFKA_INVENTORY_SCHEMA_SUBJECT",
		"OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_HEADERS", "OTEL_RESOURCE_ATTRIBUTES",
		"TRACING_SAMPLING_PROBABILITY",
	} {
		_ = viper.BindEnv(key)
	}

	viper.SetDefault("SERVICE_NAME", "inventory-transaction-service")
	viper.SetDefault("SERVER_PORT", "14330")
	viper.SetDefault("AWS_REGION", "ap-southeast-1")
	viper.SetDefault("DYNAMODB_TABLE_NAME", "inventory-transaction-service")
	viper.SetDefault("KAFKA_TOPIC_NAME", "inventory.transaction.status.updated")
	viper.SetDefault("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
	viper.SetDefault("KAFKA_AUTO_REGISTER_SCHEMAS", true)
	viper.SetDefault("KAFKA_INVENTORY_SCHEMA_SUBJECT", "inventory.transaction.status.updated")
	viper.SetDefault("TRACING_SAMPLING_PROBABILITY", 0.1)

	if err := viper.ReadInConfig(); err != nil {
		if _, notFound := err.(viper.ConfigFileNotFoundError); notFound {
			// app.env not present — try .env (Docker / Compose convention)
			viper.SetConfigFile(filepath.Join(path, ".env"))
			if err2 := viper.ReadInConfig(); err2 != nil {
				if _, notFound2 := err2.(viper.ConfigFileNotFoundError); !notFound2 {
					return cfg, fmt.Errorf("failed to read config file: %w", err2)
				}
			}
		} else {
			return cfg, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}
