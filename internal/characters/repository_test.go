package characters

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ggoulart/michael-connelly-api/internal/books"
	"github.com/ggoulart/michael-connelly-api/internal/dynamo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRepository_Save(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(*MockDynamoDBClient)
		want    Character
		wantErr error
	}{
		{
			name: "when failed to save character because already exists",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{}
				item["id"] = &types.AttributeValueMemberS{Value: ""}
				item["name"] = &types.AttributeValueMemberS{Value: "Harry Bosch"}
				item["books"] = &types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberS{Value: "book-id-1"}, &types.AttributeValueMemberS{Value: "book-id-2"}}}
				m.On("Save", ctx, "some-table-name", item, "Harry Bosch").Return("", dynamo.ErrDuplicated).Once()
				output := map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "random-id"}, "name": &types.AttributeValueMemberS{Value: "Harry Bosch"}}
				m.On("GetByUniqueKey", ctx, "some-table-name", "Harry Bosch").Return(output, nil).Once()
			},
			want: Character{ID: "random-id", Name: "Harry Bosch"},
		},
		{
			name: "when failed to save character",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{}
				item["id"] = &types.AttributeValueMemberS{Value: ""}
				item["name"] = &types.AttributeValueMemberS{Value: "Harry Bosch"}
				item["books"] = &types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberS{Value: "book-id-1"}, &types.AttributeValueMemberS{Value: "book-id-2"}}}
				m.On("Save", ctx, "some-table-name", item, "Harry Bosch").Return("", assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "when successfully saved character",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{}
				item["id"] = &types.AttributeValueMemberS{Value: ""}
				item["name"] = &types.AttributeValueMemberS{Value: "Harry Bosch"}
				item["books"] = &types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberS{Value: "book-id-1"}, &types.AttributeValueMemberS{Value: "book-id-2"}}}
				m.On("Save", ctx, "some-table-name", item, "Harry Bosch").Return("c6767b2d-438b-4d4c-8b1a-659130a640ca", nil)
			},
			want: Character{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch", Books: []books.Book{{ID: "book-id-1"}, {ID: "book-id-2"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)

			r := NewRepository(mockDynamoDBClient, "some-table-name")

			character := Character{Name: "Harry Bosch", Books: []books.Book{{ID: "book-id-1"}, {ID: "book-id-2"}}}
			got, err := r.Save(ctx, character)

			assert.Equal(t, got, tt.want)
			assert.Equal(t, tt.wantErr, err)
			mockDynamoDBClient.AssertExpectations(t)
		})
	}
}

func TestRepository_GetById(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(*MockDynamoDBClient)
		want    Character
		wantErr error
	}{
		{
			name: "when failed to get character",
			setup: func(m *MockDynamoDBClient) {
				m.On("GetByID", ctx, "some-table-name", "a-random-character-id").Return(map[string]types.AttributeValue{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "when failed to marshal output",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "character-123"}, "name": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}}}
				m.On("GetByID", ctx, "some-table-name", "a-random-character-id").Return(item, nil)
			},
			wantErr: fmt.Errorf("failed to unmarshal character: %w", &attributevalue.UnmarshalTypeError{Value: "map", Type: reflect.TypeOf("string")}),
		},
		{
			name: "when success get character",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "character-123"}}
				m.On("GetByID", ctx, "some-table-name", "a-random-character-id").Return(item, nil)
			},
			want: Character{ID: "character-123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)

			r := NewRepository(mockDynamoDBClient, "some-table-name")

			got, err := r.GetById(ctx, "a-random-character-id")

			assert.Equal(t, got, tt.want)
			assert.Equal(t, tt.wantErr, err)
			mockDynamoDBClient.AssertExpectations(t)
		})
	}
}

func TestRepository_GetByName(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(*MockDynamoDBClient)
		want    Character
		wantErr error
	}{
		{
			name: "when failed to get character",
			setup: func(m *MockDynamoDBClient) {
				m.On("GetByUniqueKey", ctx, "table-name", "Harry Bosch").Return(map[string]types.AttributeValue{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "when failed to unmarshal character",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{"name": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}}}
				m.On("GetByUniqueKey", ctx, "table-name", "Harry Bosch").Return(item, nil)
			},
			wantErr: fmt.Errorf("failed to unmarshal character: %w", &attributevalue.UnmarshalTypeError{Value: "map", Type: reflect.TypeOf("string")}),
		},
		{
			name: "when success get character",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "random-id"}, "name": &types.AttributeValueMemberS{Value: "Harry Bosch"}}
				m.On("GetByUniqueKey", ctx, "table-name", "Harry Bosch").Return(item, nil)
			},
			want: Character{ID: "random-id", Name: "Harry Bosch"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)

			r := NewRepository(mockDynamoDBClient, "table-name")

			got, err := r.GetByName(ctx, "Harry Bosch")

			assert.Equal(t, got, tt.want)
			assert.Equal(t, tt.wantErr, err)
			mockDynamoDBClient.AssertExpectations(t)
		})
	}
}

type MockDynamoDBClient struct {
	mock.Mock
}

func (m *MockDynamoDBClient) Save(ctx context.Context, tableName string, item map[string]types.AttributeValue, uniqueKey string) (string, error) {
	args := m.Called(ctx, tableName, item, uniqueKey)
	return args.String(0), args.Error(1)
}

func (m *MockDynamoDBClient) GetByID(ctx context.Context, tableName string, id string) (map[string]types.AttributeValue, error) {
	args := m.Called(ctx, tableName, id)
	return args.Get(0).(map[string]types.AttributeValue), args.Error(1)
}

func (m *MockDynamoDBClient) GetByUniqueKey(ctx context.Context, tableName string, value string) (map[string]types.AttributeValue, error) {
	args := m.Called(ctx, tableName, value)
	return args.Get(0).(map[string]types.AttributeValue), args.Error(1)
}
