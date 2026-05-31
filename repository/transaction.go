package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"inventory-transaction-service/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// ErrTransactionNotFound is returned by FindById when no item matches the
// supplied composite key.
var ErrTransactionNotFound = errors.New("transaction not found")

// TransactionRepository persists TransactionRecord items in DynamoDB.
//
// The table uses `inventoryId` as the partition key and `id` as the sort key,
// matching the Java entity annotations.
type TransactionRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewTransactionRepository(client *dynamodb.Client, tableName string) *TransactionRepository {
	return &TransactionRepository{client: client, tableName: tableName}
}

// EnsureTable creates the underlying DynamoDB table if it does not yet exist.
// Mirrors DynamoDBConfig.createTableIfNotExist from the Java service.
func (r *TransactionRepository) EnsureTable(ctx context.Context) error {
	_, err := r.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(r.tableName),
	})
	if err == nil {
		slog.Info("DynamoDB table already exists", slog.String("table", r.tableName))
		return nil
	}

	var notFound *dtypes.ResourceNotFoundException
	if !errors.As(err, &notFound) {
		return fmt.Errorf("describe table: %w", err)
	}

	slog.Info("Creating DynamoDB table", slog.String("table", r.tableName))
	_, err = r.client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(r.tableName),
		AttributeDefinitions: []dtypes.AttributeDefinition{
			{AttributeName: aws.String("inventoryId"), AttributeType: dtypes.ScalarAttributeTypeS},
			{AttributeName: aws.String("id"), AttributeType: dtypes.ScalarAttributeTypeS},
		},
		KeySchema: []dtypes.KeySchemaElement{
			{AttributeName: aws.String("inventoryId"), KeyType: dtypes.KeyTypeHash},
			{AttributeName: aws.String("id"), KeyType: dtypes.KeyTypeRange},
		},
		BillingMode: dtypes.BillingModePayPerRequest,
	})
	if err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	waiter := dynamodb.NewTableExistsWaiter(r.client)
	if err := waiter.Wait(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(r.tableName)}, 60); err != nil {
		return fmt.Errorf("wait table exists: %w", err)
	}

	slog.Info("DynamoDB table created", slog.String("table", r.tableName))
	return nil
}

// Save inserts the record (PutItem semantics — overwrites on key collision).
func (r *TransactionRepository) Save(ctx context.Context, record models.TransactionRecord) (models.TransactionRecord, error) {
	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return record, fmt.Errorf("marshal record: %w", err)
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return record, fmt.Errorf("put item: %w", err)
	}
	return record, nil
}

// FindById fetches the record with the given composite key.
// Returns ErrTransactionNotFound if the item does not exist.
func (r *TransactionRepository) FindById(ctx context.Context, inventoryId, id string) (models.TransactionRecord, error) {
	var record models.TransactionRecord

	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]dtypes.AttributeValue{
			"inventoryId": &dtypes.AttributeValueMemberS{Value: inventoryId},
			"id":          &dtypes.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return record, fmt.Errorf("get item: %w", err)
	}
	if len(out.Item) == 0 {
		return record, ErrTransactionNotFound
	}
	if err := attributevalue.UnmarshalMap(out.Item, &record); err != nil {
		return record, fmt.Errorf("unmarshal item: %w", err)
	}
	return record, nil
}

// Update replaces the existing record. Uses PutItem to match the
// "full replace" semantics of the Java enhanced-client `updateItem`.
func (r *TransactionRepository) Update(ctx context.Context, record models.TransactionRecord) (models.TransactionRecord, error) {
	return r.Save(ctx, record)
}
