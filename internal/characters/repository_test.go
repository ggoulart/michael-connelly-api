package characters

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRepository_Save(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockDynamoDBClient)
		want    Character
		wantErr error
	}{
		{
			name: "when failed to save character",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{}
				item["Id"] = &types.AttributeValueMemberS{Value: "c6767b2d-438b-4d4c-8b1a-659130a640ca"}
				item["Name"] = &types.AttributeValueMemberS{Value: "Harry Bosch"}
				item["Books"] = &types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberS{Value: "book-id-1"}, &types.AttributeValueMemberS{Value: "book-id-2"}}}

				input := &dynamodb.PutItemInput{Item: item, TableName: aws.String("some-table-name")}
				m.On("PutItem", mock.AnythingOfType("backgroundCtx"), input, mock.Anything).Return(&dynamodb.PutItemOutput{}, errors.New("dynamodb.PutItem error"))
			},
			want:    Character{},
			wantErr: fmt.Errorf("%w. failed to save character: %w", ErrDynamodb, errors.New("dynamodb.PutItem error")),
		},
		{
			name: "when successfully saved character",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{}
				item["Id"] = &types.AttributeValueMemberS{Value: "c6767b2d-438b-4d4c-8b1a-659130a640ca"}
				item["Name"] = &types.AttributeValueMemberS{Value: "Harry Bosch"}
				item["Books"] = &types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberS{Value: "book-id-1"}, &types.AttributeValueMemberS{Value: "book-id-2"}}}

				input := &dynamodb.PutItemInput{Item: item, TableName: aws.String("some-table-name")}
				m.On("PutItem", mock.AnythingOfType("backgroundCtx"), input, mock.Anything).Return(&dynamodb.PutItemOutput{}, nil)
			},
			want: Character{Id: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch", Books: []string{"book-id-1", "book-id-2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)

			r := NewRepository(mockDynamoDBClient, "some-table-name", func() uuid.UUID {
				return uuid.MustParse("c6767b2d-438b-4d4c-8b1a-659130a640ca")
			})

			character := Character{Name: "Harry Bosch", Books: []string{"book-id-1", "book-id-2"}}
			got, err := r.Save(context.Background(), character)

			assert.Equal(t, got, tt.want)
			assert.Equal(t, tt.wantErr, err)
			mockDynamoDBClient.AssertExpectations(t)
		})
	}
}

func TestRepository_GetById(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockDynamoDBClient)
		want    Character
		wantErr error
	}{
		{
			name: "when failed to get character",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("some-table-name"), Key: map[string]types.AttributeValue{"Id": &types.AttributeValueMemberS{Value: "a-random-character-id"}}}
				m.On("GetItem", mock.AnythingOfType("backgroundCtx"), input, mock.Anything).Return(&dynamodb.GetItemOutput{}, assert.AnError)
			},
			wantErr: fmt.Errorf("%w. failed to fetch character, id: %s, err: %w", ErrDynamodb, "a-random-character-id", assert.AnError),
		},
		{
			name: "when character not found",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("some-table-name"), Key: map[string]types.AttributeValue{"Id": &types.AttributeValueMemberS{Value: "a-random-character-id"}}}
				m.On("GetItem", mock.AnythingOfType("backgroundCtx"), input, mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)
			},
			wantErr: fmt.Errorf("%w. character id: %s", ErrNotFound, "a-random-character-id"),
		},
		{
			name: "when failed to marshal output",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("some-table-name"), Key: map[string]types.AttributeValue{"Id": &types.AttributeValueMemberS{Value: "a-random-character-id"}}}
				output := &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{
					"id":   &types.AttributeValueMemberS{Value: "character-123"},
					"name": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}},
				}}
				m.On("GetItem", mock.AnythingOfType("backgroundCtx"), input, mock.Anything).Return(output, nil)
			},
			wantErr: fmt.Errorf("%w. failed to unmarshal character: %w", ErrDynamodb, errors.New("unmarshal failed, cannot unmarshal map into Go value type string")),
		},
		{
			name: "when success get character",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("some-table-name"), Key: map[string]types.AttributeValue{"Id": &types.AttributeValueMemberS{Value: "a-random-character-id"}}}
				output := &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "character-123"}}}
				m.On("GetItem", mock.AnythingOfType("backgroundCtx"), input, mock.Anything).Return(output, nil)
			},
			want: Character{Id: "character-123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)

			r := NewRepository(mockDynamoDBClient, "some-table-name", nil)

			got, err := r.GetById(context.Background(), "a-random-character-id")

			assert.Equal(t, got, tt.want)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}
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
