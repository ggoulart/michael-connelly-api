package books

import (
	"context"
	"errors"
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
				item["id"] = &types.AttributeValueMemberS{Value: "c6767b2d-438b-4d4c-8b1a-659130a640ca"}
				item["title"] = &types.AttributeValueMemberS{Value: "The Black Echo"}
				item["year"] = &types.AttributeValueMemberN{Value: "1992"}
				item["blurb"] = &types.AttributeValueMemberS{Value: "a random blurb"}

				input := &dynamodb.PutItemInput{Item: item, TableName: aws.String("some-table-name")}
				m.On("PutItem", mock.AnythingOfType("backgroundCtx"), input, mock.Anything).Return(&dynamodb.PutItemOutput{}, errors.New("dynamodb.PutItem error"))
			},
			want:    Book{},
			wantErr: fmt.Errorf("%w. failed to save book: %w", ErrDynamodb, errors.New("dynamodb.PutItem error")),
		},
		{
			name: "when successfully saved book",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{}
				item["id"] = &types.AttributeValueMemberS{Value: "c6767b2d-438b-4d4c-8b1a-659130a640ca"}
				item["title"] = &types.AttributeValueMemberS{Value: "The Black Echo"}
				item["year"] = &types.AttributeValueMemberN{Value: "1992"}
				item["blurb"] = &types.AttributeValueMemberS{Value: "a random blurb"}

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
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(*MockDynamoDBClient)
		want    Book
		wantErr error
	}{
		{
			name: "when failed to get book",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("some-table-name"), Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "a-random-book-id"}}}
				m.On("GetItem", ctx, input, mock.Anything).Return(&dynamodb.GetItemOutput{}, assert.AnError)
			},
			wantErr: fmt.Errorf("%w. failed to fetch book, id: %s, err: %w", ErrDynamodb, "a-random-book-id", assert.AnError),
		},
		{
			name: "when book not found",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("some-table-name"), Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "a-random-book-id"}}}
				m.On("GetItem", ctx, input, mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)
			},
			wantErr: fmt.Errorf("%w. book id: %s", ErrNotFound, "a-random-book-id"),
		},
		{
			name: "when failed to marshal output",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("some-table-name"), Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "a-random-book-id"}}}
				output := &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{
					"id":    &types.AttributeValueMemberS{Value: "book-123"},
					"title": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}},
				}}
				m.On("GetItem", ctx, input, mock.Anything).Return(output, nil)
			},
			wantErr: fmt.Errorf("%w. failed to unmarshal book: %w", ErrDynamodb, &attributevalue.UnmarshalTypeError{Value: "map", Type: reflect.TypeOf("string")}),
		},
		{
			name: "when success get book",
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.GetItemInput{TableName: aws.String("some-table-name"), Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "a-random-book-id"}}}
				output := &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "book-123"}}}
				m.On("GetItem", ctx, input, mock.Anything).Return(output, nil)
			},
			want: Book{Id: "book-123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)

			r := NewRepository(mockDynamoDBClient, "some-table-name", nil)

			got, err := r.GetById(ctx, "a-random-book-id")

			assert.Equal(t, got, tt.want)
			assert.Equal(t, tt.wantErr, err)
			mockDynamoDBClient.AssertExpectations(t)
		})
	}
}

func TestRepository_GetByNames(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name       string
		bookTitles []string
		setup      func(*MockDynamoDBClient)
		want       []Book
		wantErr    error
	}{
		{
			name:       "when book titles is empty",
			bookTitles: []string{},
			setup:      func(*MockDynamoDBClient) {},
		},
		{
			name:       "when failed to query book",
			bookTitles: []string{"The Black Echo"},
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.QueryInput{TableName: aws.String("some-table-name"), IndexName: aws.String("books_title"), KeyConditionExpression: aws.String("title = :title"), ExpressionAttributeValues: map[string]types.AttributeValue{":title": &types.AttributeValueMemberS{Value: "The Black Echo"}}}
				m.On("Query", ctx, input, mock.Anything).Return(&dynamodb.QueryOutput{}, assert.AnError).Once()
			},
			wantErr: fmt.Errorf("%w. failed to fetch book, title: %s, err: %w", ErrDynamodb, "The Black Echo", assert.AnError),
		},
		{
			name:       "when failed to query second book",
			bookTitles: []string{"The Black Echo", "The Black Ice"},
			setup: func(m *MockDynamoDBClient) {
				firstInput := &dynamodb.QueryInput{TableName: aws.String("some-table-name"), IndexName: aws.String("books_title"), KeyConditionExpression: aws.String("title = :title"), ExpressionAttributeValues: map[string]types.AttributeValue{":title": &types.AttributeValueMemberS{Value: "The Black Echo"}}}
				firstOutput := &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{{"id": &types.AttributeValueMemberS{Value: "a-random-book-id"}}}}
				m.On("Query", ctx, firstInput, mock.Anything).Return(firstOutput, nil).Once()
				secondInput := &dynamodb.QueryInput{TableName: aws.String("some-table-name"), IndexName: aws.String("books_title"), KeyConditionExpression: aws.String("title = :title"), ExpressionAttributeValues: map[string]types.AttributeValue{":title": &types.AttributeValueMemberS{Value: "The Black Ice"}}}
				m.On("Query", ctx, secondInput, mock.Anything).Return(&dynamodb.QueryOutput{}, assert.AnError).Once()
			},
			wantErr: fmt.Errorf("%w. failed to fetch book, title: %s, err: %w", ErrDynamodb, "The Black Ice", assert.AnError),
		},
		{
			name:       "when failed to marshal output",
			bookTitles: []string{"The Black Echo"},
			setup: func(m *MockDynamoDBClient) {
				input := &dynamodb.QueryInput{TableName: aws.String("some-table-name"), IndexName: aws.String("books_title"), KeyConditionExpression: aws.String("title = :title"), ExpressionAttributeValues: map[string]types.AttributeValue{":title": &types.AttributeValueMemberS{Value: "The Black Echo"}}}
				output := &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{{"title": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}}}}}
				m.On("Query", ctx, input, mock.Anything).Return(output, nil).Once()
			},
			wantErr: fmt.Errorf("%w. failed to unmarshal book: %w", ErrDynamodb, &attributevalue.UnmarshalTypeError{Value: "map", Type: reflect.TypeOf("The Black Echo")}),
		},
		{
			name:       "when successfully get books",
			bookTitles: []string{"The Black Echo", "The Black Ice"},
			setup: func(m *MockDynamoDBClient) {
				firstInput := &dynamodb.QueryInput{TableName: aws.String("some-table-name"), IndexName: aws.String("books_title"), KeyConditionExpression: aws.String("title = :title"), ExpressionAttributeValues: map[string]types.AttributeValue{":title": &types.AttributeValueMemberS{Value: "The Black Echo"}}}
				firstOutput := &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{{"id": &types.AttributeValueMemberS{Value: "the-black-echo-id"}, "title": &types.AttributeValueMemberS{Value: "The Black Echo"}}}}
				m.On("Query", ctx, firstInput, mock.Anything).Return(firstOutput, nil).Once()
				secondInput := &dynamodb.QueryInput{TableName: aws.String("some-table-name"), IndexName: aws.String("books_title"), KeyConditionExpression: aws.String("title = :title"), ExpressionAttributeValues: map[string]types.AttributeValue{":title": &types.AttributeValueMemberS{Value: "The Black Ice"}}}
				secondOutput := &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{{"id": &types.AttributeValueMemberS{Value: "the-black-ice-id"}, "title": &types.AttributeValueMemberS{Value: "The Black Ice"}}}}
				m.On("Query", ctx, secondInput, mock.Anything).Return(secondOutput, nil).Once()
			},
			want: []Book{{Id: "the-black-echo-id", Title: "The Black Echo"}, {Id: "the-black-ice-id", Title: "The Black Ice"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)

			r := NewRepository(mockDynamoDBClient, "some-table-name", nil)

			got, err := r.GetByNames(ctx, tt.bookTitles)

			assert.Equal(t, got, tt.want)
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

func (m *MockDynamoDBClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}
