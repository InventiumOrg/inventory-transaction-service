package repository

import (
	"context"
	"fmt"

	"inventory-transaction-service/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// NewDynamoDBClient builds a DynamoDB client honoring an optional
// `DYNAMODB_ENDPOINT` override (used for DynamoDB Local in dev/tests).
func NewDynamoDBClient(ctx context.Context, cfg config.Config) (*dynamodb.Client, error) {
	loadOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.AWSRegion),
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	var clientOpts []func(*dynamodb.Options)
	if cfg.DynamoDBEndpoint != "" {
		endpoint := cfg.DynamoDBEndpoint
		clientOpts = append(clientOpts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	return dynamodb.NewFromConfig(awsCfg, clientOpts...), nil
}
