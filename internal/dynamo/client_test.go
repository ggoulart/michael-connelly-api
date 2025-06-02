package dynamo

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_Save(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(*MockDynamoDBClient)
		want    string
		wantErr error
	}{
		{
			name: "when failed to save because unique key already exists",
			setup: func(m *MockDynamoDBClient) {
				uniqueKeyItem := map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "table-name#uniqueValue"}, "table_id": &types.AttributeValueMemberS{Value: "c6767b2d-438b-4d4c-8b1a-659130a640ca"}}
				item := map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "c6767b2d-438b-4d4c-8b1a-659130a640ca"}}
				input := &dynamodb.TransactWriteItemsInput{TransactItems: []types.TransactWriteItem{
					{Put: &types.Put{TableName: aws.String("unique_keys"), Item: uniqueKeyItem, ConditionExpression: aws.String("attribute_not_exists(id)")}},
					{Put: &types.Put{TableName: aws.String("table-name"), Item: item}},
				}}
				err := types.TransactionCanceledException{CancellationReasons: []types.CancellationReason{{Code: aws.String("ConditionalCheckFailed")}}}
				m.On("TransactWriteItems", ctx, input, mock.Anything).Return(&dynamodb.TransactWriteItemsOutput{}, &err).Once()
			},
			wantErr: ErrDuplicated,
		},
		{
			name: "when failed to save",
			setup: func(m *MockDynamoDBClient) {
				uniqueKeyItem := map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "table-name#uniqueValue"}, "table_id": &types.AttributeValueMemberS{Value: "c6767b2d-438b-4d4c-8b1a-659130a640ca"}}
				item := map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "c6767b2d-438b-4d4c-8b1a-659130a640ca"}}
				input := &dynamodb.TransactWriteItemsInput{TransactItems: []types.TransactWriteItem{
					{Put: &types.Put{TableName: aws.String("unique_keys"), Item: uniqueKeyItem, ConditionExpression: aws.String("attribute_not_exists(id)")}},
					{Put: &types.Put{TableName: aws.String("table-name"), Item: item}},
				}}
				m.On("TransactWriteItems", ctx, input, mock.Anything).Return(&dynamodb.TransactWriteItemsOutput{}, assert.AnError).Once()
			},
			wantErr: fmt.Errorf("%w. failed to save character: %w", ErrDynamodb, assert.AnError),
		},
		{
			name: "when successfully saved",
			setup: func(m *MockDynamoDBClient) {
				uniqueKeyItem := map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "table-name#uniqueValue"}, "table_id": &types.AttributeValueMemberS{Value: "c6767b2d-438b-4d4c-8b1a-659130a640ca"}}
				item := map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "c6767b2d-438b-4d4c-8b1a-659130a640ca"}}
				input := &dynamodb.TransactWriteItemsInput{TransactItems: []types.TransactWriteItem{
					{Put: &types.Put{TableName: aws.String("unique_keys"), Item: uniqueKeyItem, ConditionExpression: aws.String("attribute_not_exists(id)")}},
					{Put: &types.Put{TableName: aws.String("table-name"), Item: item}},
				}}
				m.On("TransactWriteItems", ctx, input, mock.Anything).Return(&dynamodb.TransactWriteItemsOutput{}, nil).Once()
			},
			want: "c6767b2d-438b-4d4c-8b1a-659130a640ca",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)
			c := NewClient(mockDynamoDBClient, func() uuid.UUID { return uuid.MustParse("c6767b2d-438b-4d4c-8b1a-659130a640ca") })

			item := map[string]types.AttributeValue{}
			got, err := c.Save(ctx, "table-name", item, "uniqueValue")

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
			mockDynamoDBClient.AssertExpectations(t)
		})
	}
}

func TestClient_GetByID(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(*MockDynamoDBClient)
		want    map[string]types.AttributeValue
		wantErr error
	}{
		{
			name: "when failed to get by id",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("table-name"), Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "random-id"}}}
				m.On("GetItem", ctx, input, mock.Anything).Return(&dynamodb.GetItemOutput{}, assert.AnError).Once()
			},
			wantErr: fmt.Errorf("%w. failed to get item id: %s from table: %s. err: %w", ErrDynamodb, "random-id", "table-name", assert.AnError),
		},
		{
			name: "when id not found",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("table-name"), Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "random-id"}}}
				output := &dynamodb.GetItemOutput{}
				m.On("GetItem", ctx, input, mock.Anything).Return(output, nil).Once()
			},
			wantErr: fmt.Errorf("%w. id: %s", ErrNotFound, "random-id"),
		},
		{
			name: "when successfully get by id",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("table-name"), Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "random-id"}}}
				output := &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "random-id"}}}
				m.On("GetItem", ctx, input, mock.Anything).Return(output, nil).Once()
			},
			want: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "random-id"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)
			c := NewClient(mockDynamoDBClient, nil)

			got, err := c.GetByID(ctx, "table-name", "random-id")

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
			mockDynamoDBClient.AssertExpectations(t)
		})
	}
}

func TestClient_GetByUniqueKey(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(*MockDynamoDBClient)
		want    map[string]types.AttributeValue
		wantErr error
	}{
		{
			name: "when failed to get id by unique key",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("unique_keys"), Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "table-name#Harry Bosch"}}}
				m.On("GetItem", ctx, input, mock.Anything).Return(&dynamodb.GetItemOutput{}, assert.AnError).Once()
			},
			wantErr: fmt.Errorf("%w. failed to get item id: %s from table: %s. err: %w", ErrDynamodb, "table-name#Harry Bosch", "unique_keys", assert.AnError),
		},
		{
			name: "when unique key not found",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("unique_keys"), Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "table-name#Harry Bosch"}}}
				m.On("GetItem", ctx, input, mock.Anything).Return(&dynamodb.GetItemOutput{}, nil).Once()
			},
			wantErr: fmt.Errorf("%w. id: %s", ErrNotFound, "table-name#Harry Bosch"),
		},
		{
			name: "when failed to unmarshal item",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("unique_keys"), Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "table-name#Harry Bosch"}}}
				output := &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{"id": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}}}}
				m.On("GetItem", ctx, input, mock.Anything).Return(output, nil).Once()
			},
			wantErr: fmt.Errorf("%w. failed to unmarshal. table: %s, value: %s. err: %w", ErrDynamodb, "table-name", "Harry Bosch", &attributevalue.UnmarshalTypeError{Value: "map", Type: reflect.TypeOf("string")}),
		},
		{
			name: "when failed to get item",
			setup: func(m *MockDynamoDBClient) {
				ukInput := &dynamodb.GetItemInput{TableName: aws.String("unique_keys"), Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "table-name#Harry Bosch"}}}
				ukOutput := &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "table-name#Harry Bosch"}, "table_id": &types.AttributeValueMemberS{Value: "random-id"}}}
				m.On("GetItem", ctx, ukInput, mock.Anything).Return(ukOutput, nil).Once()
				itemInput := &dynamodb.GetItemInput{TableName: aws.String("table-name"), Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "random-id"}}}
				m.On("GetItem", ctx, itemInput, mock.Anything).Return(&dynamodb.GetItemOutput{}, assert.AnError).Once()
			},
			wantErr: fmt.Errorf("%w. failed to get item id: %s from table: %s. err: %w", ErrDynamodb, "random-id", "table-name", assert.AnError),
		},
		{
			name: "when successfully get by unique key",
			setup: func(m *MockDynamoDBClient) {
				ukInput := &dynamodb.GetItemInput{TableName: aws.String("unique_keys"), Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "table-name#Harry Bosch"}}}
				ukOutput := &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "table-name#Harry Bosch"}, "table_id": &types.AttributeValueMemberS{Value: "random-id"}}}
				m.On("GetItem", ctx, ukInput, mock.Anything).Return(ukOutput, nil).Once()
				itemInput := &dynamodb.GetItemInput{TableName: aws.String("table-name"), Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "random-id"}}}
				itemOutput := &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "random-id"}}}
				m.On("GetItem", ctx, itemInput, mock.Anything).Return(itemOutput, nil).Once()
			},
			want: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "random-id"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)
			c := NewClient(mockDynamoDBClient, nil)

			got, err := c.GetByUniqueKey(ctx, "table-name", "Harry Bosch")

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
			mockDynamoDBClient.AssertExpectations(t)
		})
	}
}

type MockDynamoDBClient struct {
	mock.Mock
}

func (m *MockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.UpdateItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}

func (m *MockDynamoDBClient) TransactWriteItems(ctx context.Context, params *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.TransactWriteItemsOutput), args.Error(1)
}
