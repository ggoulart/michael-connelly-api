package dynamo

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

var ErrDynamodb = errors.New("dynamodb: error")
var ErrNotFound = errors.New("dynamodb: not found")
var ErrDuplicated = errors.New("dynamodb: duplicated")

type Dynamodb interface {
	GetItem(ctx context.Context, input *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	TransactWriteItems(ctx context.Context, input *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
	CreateTable(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error)
	ListTables(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error)
}

var uniqueKeyTable = "unique_keys"

type Client struct {
	dynamoDB Dynamodb
	uuidGen  func() uuid.UUID
}

func NewClient(dynamodb Dynamodb, uuidGen func() uuid.UUID) *Client {
	return &Client{dynamoDB: dynamodb, uuidGen: uuidGen}
}

func (c *Client) Save(ctx context.Context, tableName string, item map[string]types.AttributeValue, uniqueValue string) (string, error) {
	tableID := c.uuidGen().String()
	item["id"] = &types.AttributeValueMemberS{Value: tableID}

	uniqueTableID := fmt.Sprintf("%s#%s", tableName, uniqueValue)
	uniqueKeyItem := map[string]types.AttributeValue{
		"id":       &types.AttributeValueMemberS{Value: uniqueTableID},
		"table_id": &types.AttributeValueMemberS{Value: tableID},
	}

	_, err := c.dynamoDB.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{Put: &types.Put{TableName: aws.String(uniqueKeyTable), Item: uniqueKeyItem, ConditionExpression: aws.String("attribute_not_exists(id)")}},
			{Put: &types.Put{TableName: aws.String(tableName), Item: item}},
		},
	})

	var tce *types.TransactionCanceledException
	if errors.As(err, &tce) {
		if len(tce.CancellationReasons) > 0 && tce.CancellationReasons[0].Code != nil {
			if *tce.CancellationReasons[0].Code == "ConditionalCheckFailed" {
				return "", ErrDuplicated
			}
		}
	}
	if err != nil {
		return "", fmt.Errorf("%w. failed to save character: %w", ErrDynamodb, err)
	}

	return tableID, nil
}

func (c *Client) GetByID(ctx context.Context, tableName string, id string) (map[string]types.AttributeValue, error) {
	output, err := c.dynamoDB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: id}},
	})
	if err != nil {
		return nil, fmt.Errorf("%w. failed to get item id: %s from table: %s. err: %w", ErrDynamodb, id, tableName, err)
	}

	if output.Item == nil {
		return nil, fmt.Errorf("%w. id: %s", ErrNotFound, id)
	}

	return output.Item, nil
}

func (c *Client) GetByUniqueKey(ctx context.Context, tableName string, value string) (map[string]types.AttributeValue, error) {
	ukItem, err := c.GetByID(ctx, uniqueKeyTable, fmt.Sprintf("%s#%s", tableName, value))
	if err != nil {
		return nil, err
	}

	var uniqueKeys UniqueKeys
	err = attributevalue.UnmarshalMap(ukItem, &uniqueKeys)
	if err != nil {
		return nil, fmt.Errorf("%w. failed to unmarshal. table: %s, value: %s. err: %w", ErrDynamodb, tableName, value, err)
	}

	item, err := c.GetByID(ctx, tableName, uniqueKeys.TableID)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (c *Client) CreateTables(ctx context.Context) error {
	tables := []struct {
		Name     string
		HashKey  string
		HashType types.ScalarAttributeType
	}{
		{"unique_keys", "id", types.ScalarAttributeTypeS},
		{"books", "id", types.ScalarAttributeTypeS},
		{"characters", "id", types.ScalarAttributeTypeS},
		{"series", "id", types.ScalarAttributeTypeS},
	}

	for _, tbl := range tables {
		_, err := c.dynamoDB.CreateTable(ctx, &dynamodb.CreateTableInput{
			TableName:            aws.String(tbl.Name),
			AttributeDefinitions: []types.AttributeDefinition{{AttributeName: aws.String(tbl.HashKey), AttributeType: tbl.HashType}},
			KeySchema:            []types.KeySchemaElement{{AttributeName: aws.String(tbl.HashKey), KeyType: types.KeyTypeHash}},
			BillingMode:          types.BillingModePayPerRequest,
		})
		if err != nil {
			var resourceInUse *types.ResourceInUseException
			if errors.As(err, &resourceInUse) {
				continue
			}
			return fmt.Errorf("failed to create table %s: %w", tbl.Name, err)
		}
	}

	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	var limit int32 = 1
	_, err := c.dynamoDB.ListTables(ctx, &dynamodb.ListTablesInput{Limit: &limit})
	if err != nil {
		return fmt.Errorf("%w. failed to ping dynamodb: %w", ErrDynamodb, err)
	}

	return nil
}

type UniqueKeys struct {
	ID      string `dynamodbav:"id"`
	TableID string `dynamodbav:"table_id"`
}
