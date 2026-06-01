package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all runtime configuration for the inventory-transaction-service.
type Config struct {
	ServiceName string `mapstructure:"SERVICE_NAME"`
	ServerPort  string `mapstructure:"SERVER_PORT"`

	// AWS / DynamoDB
	AWSRegion         string `mapstructure:"AWS_REGION"`
	DynamoDBEndpoint  string `mapstructure:"DYNAMODB_ENDPOINT"`
	DynamoDBTableName string `mapstructure:"DYNAMODB_TABLE_NAME"`

	// Kafka
	KafkaTopicName       string `mapstructure:"KAFKA_TOPIC_NAME"`
	KafkaBootstrapServer string `mapstructure:"KAFKA_BOOTSTRAP_SERVERS"`
	KafkaCAFilePath      string `mapstructure:"KAFKA_CA_FILE_PATH"`
	KafkaUsername        string `mapstructure:"KAFKA_USERNAME"`
	KafkaPassword        string `mapstructure:"KAFKA_PASSWORD"`
	KafkaAutoRegister    bool   `mapstructure:"KAFKA_AUTO_REGISTER_SCHEMAS"`

	// Schema Registry
	SchemaRegistryURL      string `mapstructure:"SCHEMA_REGISTRY_URL"`
	SchemaRegistryUsername string `mapstructure:"SCHEMA_REGISTRY_USERNAME"`
	SchemaRegistryPassword string `mapstructure:"SCHEMA_REGISTRY_PASSWORD"`
	SchemaSubject          string `mapstructure:"KAFKA_INVENTORY_SCHEMA_SUBJECT"`

	// Observability
	OTELExporterOTLPEndpoint string  `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTELExporterOTLPHeaders  string  `mapstructure:"OTEL_EXPORTER_OTLP_HEADERS"`
	OTELResourceAttributes   string  `mapstructure:"OTEL_RESOURCE_ATTRIBUTES"`
	TracingSamplingRatio     float64 `mapstructure:"TRACING_SAMPLING_PROBABILITY"`
}

// LoadConfig loads configuration from environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AutomaticEnv()

	_ = viper.BindEnv("SERVICE_NAME")
	_ = viper.BindEnv("SERVER_PORT")
	_ = viper.BindEnv("AWS_REGION")
	_ = viper.BindEnv("DYNAMODB_ENDPOINT")
	_ = viper.BindEnv("DYNAMODB_TABLE_NAME")
	_ = viper.BindEnv("KAFKA_TOPIC_NAME")
	_ = viper.BindEnv("KAFKA_BOOTSTRAP_SERVERS")
	_ = viper.BindEnv("KAFKA_CA_FILE_PATH")
	_ = viper.BindEnv("KAFKA_USERNAME")
	_ = viper.BindEnv("KAFKA_PASSWORD")
	_ = viper.BindEnv("KAFKA_AUTO_REGISTER_SCHEMAS")
	_ = viper.BindEnv("SCHEMA_REGISTRY_URL")
	_ = viper.BindEnv("SCHEMA_REGISTRY_USERNAME")
	_ = viper.BindEnv("SCHEMA_REGISTRY_PASSWORD")
	_ = viper.BindEnv("KAFKA_INVENTORY_SCHEMA_SUBJECT")
	_ = viper.BindEnv("OTEL_EXPORTER_OTLP_ENDPOINT")
	_ = viper.BindEnv("OTEL_EXPORTER_OTLP_HEADERS")
	_ = viper.BindEnv("OTEL_RESOURCE_ATTRIBUTES")
	_ = viper.BindEnv("TRACING_SAMPLING_PROBABILITY")

	viper.SetDefault("SERVICE_NAME", "inventory-transaction-service")
	viper.SetDefault("SERVER_PORT", "14330")
	viper.SetDefault("AWS_REGION", "ap-southeast-1")
	viper.SetDefault("DYNAMODB_TABLE_NAME", "inventory-transaction-service")
	viper.SetDefault("KAFKA_TOPIC_NAME", "inventory.transaction.status.updated")
	viper.SetDefault("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
	viper.SetDefault("KAFKA_AUTO_REGISTER_SCHEMAS", true)
	viper.SetDefault("KAFKA_INVENTORY_SCHEMA_SUBJECT", "inventory.transaction.status.updated")
	viper.SetDefault("TRACING_SAMPLING_PROBABILITY", 0.1)

	err = viper.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}
