package books

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
		want    Book
		wantErr error
	}{
		{
			name: "when failed to save book",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{}
				item["Id"] = &types.AttributeValueMemberS{Value: "c6767b2d-438b-4d4c-8b1a-659130a640ca"}
				item["Title"] = &types.AttributeValueMemberS{Value: "The Black Echo"}
				item["Year"] = &types.AttributeValueMemberN{Value: "1992"}
				item["Blurb"] = &types.AttributeValueMemberS{Value: "a random blurb"}

				input := &dynamodb.PutItemInput{Item: item, TableName: aws.String("some-table-name")}
				m.On("PutItem", mock.AnythingOfType("backgroundCtx"), input, mock.Anything).Return(&dynamodb.PutItemOutput{}, errors.New("dynamodb.PutItem error"))
			},
			want:    Book{},
			wantErr: fmt.Errorf("failed to save book: %w", errors.New("dynamodb.PutItem error")),
		},
		{
			name: "when successfully saved book",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{}
				item["Id"] = &types.AttributeValueMemberS{Value: "c6767b2d-438b-4d4c-8b1a-659130a640ca"}
				item["Title"] = &types.AttributeValueMemberS{Value: "The Black Echo"}
				item["Year"] = &types.AttributeValueMemberN{Value: "1992"}
				item["Blurb"] = &types.AttributeValueMemberS{Value: "a random blurb"}

				input := &dynamodb.PutItemInput{Item: item, TableName: aws.String("some-table-name")}
				m.On("PutItem", mock.AnythingOfType("backgroundCtx"), input, mock.Anything).Return(&dynamodb.PutItemOutput{}, nil)
			},
			want: Book{Id: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)

			r := NewRepository(mockDynamoDBClient, "some-table-name", func() uuid.UUID {
				return uuid.MustParse("c6767b2d-438b-4d4c-8b1a-659130a640ca")
			})

			book := Book{Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"}
			got, err := r.Save(context.Background(), book)

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
		want    Book
		wantErr error
	}{
		{
			name: "when failed to get book",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("some-table-name"), Key: map[string]types.AttributeValue{"Id": &types.AttributeValueMemberS{Value: "a-random-book-id"}}}
				m.On("GetItem", mock.AnythingOfType("backgroundCtx"), input, mock.Anything).Return(&dynamodb.GetItemOutput{}, assert.AnError)
			},
			wantErr: fmt.Errorf("failed to fetch book, id: %s, err: %w", "a-random-book-id", assert.AnError),
		},
		{
			name: "when book not found",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("some-table-name"), Key: map[string]types.AttributeValue{"Id": &types.AttributeValueMemberS{Value: "a-random-book-id"}}}
				m.On("GetItem", mock.AnythingOfType("backgroundCtx"), input, mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)
			},
			wantErr: fmt.Errorf("book not found id: %s", "a-random-book-id"),
		},
		{
			name: "when failed to marshal output",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("some-table-name"), Key: map[string]types.AttributeValue{"Id": &types.AttributeValueMemberS{Value: "a-random-book-id"}}}
				output := &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{
					"id":    &types.AttributeValueMemberS{Value: "book-123"},
					"title": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}},
				}}
				m.On("GetItem", mock.AnythingOfType("backgroundCtx"), input, mock.Anything).Return(output, nil)
			},
			wantErr: fmt.Errorf("failed to unmarshal book: %w", errors.New("unmarshal failed, cannot unmarshal map into Go value type string")),
		},
		{
			name: "when success get book",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("some-table-name"), Key: map[string]types.AttributeValue{"Id": &types.AttributeValueMemberS{Value: "a-random-book-id"}}}
				output := &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "book-123"}}}
				m.On("GetItem", mock.AnythingOfType("backgroundCtx"), input, mock.Anything).Return(output, nil)
			},
			want: Book{Id: "book-123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)

			r := NewRepository(mockDynamoDBClient, "some-table-name", nil)

			got, err := r.GetById(context.Background(), "a-random-book-id")

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
