package series

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
		want    Series
		wantErr error
	}{
		{
			name: "when failed to save series because already exists",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{
					"id":    &types.AttributeValueMemberS{Value: ""},
					"title": &types.AttributeValueMemberS{Value: "Harry Bosch"},
					"booksOrder": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
							"book_id": &types.AttributeValueMemberS{Value: "book-id-1"},
							"order":   &types.AttributeValueMemberN{Value: "1"},
						}},
					}},
				}
				m.On("Save", ctx, "series-table", item, "Harry Bosch").Return("", dynamo.ErrDuplicated).Once()
				output := map[string]types.AttributeValue{
					"id":    &types.AttributeValueMemberS{Value: "series-id-1"},
					"title": &types.AttributeValueMemberS{Value: "Harry Bosch"},
					"booksOrder": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
							"book_id": &types.AttributeValueMemberS{Value: "book-id-1"},
							"order":   &types.AttributeValueMemberN{Value: "1"},
						}},
					}},
				}
				m.On("GetByUniqueKey", ctx, "series-table", "Harry Bosch").Return(output, nil).Once()
			},
			want: Series{ID: "series-id-1", Title: "Harry Bosch", Books: []BooksOrder{{Order: 1, Book: books.Book{ID: "book-id-1"}}}},
		},
		{
			name: "when failed to save series",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{
					"id":    &types.AttributeValueMemberS{Value: ""},
					"title": &types.AttributeValueMemberS{Value: "Harry Bosch"},
					"booksOrder": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
							"book_id": &types.AttributeValueMemberS{Value: "book-id-1"},
							"order":   &types.AttributeValueMemberN{Value: "1"},
						}},
					}},
				}
				m.On("Save", ctx, "series-table", item, "Harry Bosch").Return("", assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "when successfully saved series",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{
					"id":    &types.AttributeValueMemberS{Value: ""},
					"title": &types.AttributeValueMemberS{Value: "Harry Bosch"},
					"booksOrder": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
							"book_id": &types.AttributeValueMemberS{Value: "book-id-1"},
							"order":   &types.AttributeValueMemberN{Value: "1"},
						}},
					}},
				}
				m.On("Save", ctx, "series-table", item, "Harry Bosch").Return("series-id-1", nil).Once()
			},
			want: Series{ID: "series-id-1", Title: "Harry Bosch", Books: []BooksOrder{{Order: 1, Book: books.Book{ID: "book-id-1", Title: "The Black Echo"}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)

			r := NewRepository(mockDynamoDBClient, "series-table")

			got, err := r.Save(ctx, Series{
				Title: "Harry Bosch",
				Books: []BooksOrder{{Order: 1, Book: books.Book{ID: "book-id-1", Title: "The Black Echo"}}},
			})

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
			mockDynamoDBClient.AssertExpectations(t)
		})
	}
}

func TestRepository_GetByTitle(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(*MockDynamoDBClient)
		want    Series
		wantErr error
	}{
		{
			name: "when failed to get series by title",
			setup: func(m *MockDynamoDBClient) {
				m.On("GetByUniqueKey", ctx, "series-table", "Harry Bosch").Return(map[string]types.AttributeValue{}, assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "when failed to unmarshal series",
			setup: func(m *MockDynamoDBClient) {
				output := map[string]types.AttributeValue{"title": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}}}
				m.On("GetByUniqueKey", ctx, "series-table", "Harry Bosch").Return(output, nil).Once()
			},
			wantErr: fmt.Errorf("failed to unmarshal series: %w", &attributevalue.UnmarshalTypeError{Value: "map", Type: reflect.TypeOf("string")}),
		},
		{
			name: "when successfully get series by title",
			setup: func(m *MockDynamoDBClient) {
				output := map[string]types.AttributeValue{
					"id":    &types.AttributeValueMemberS{Value: "series-id-1"},
					"title": &types.AttributeValueMemberS{Value: "Harry Bosch"},
					"booksOrder": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
							"book_id": &types.AttributeValueMemberS{Value: "book-id-1"},
							"order":   &types.AttributeValueMemberN{Value: "1"},
						}},
					}},
				}
				m.On("GetByUniqueKey", ctx, "series-table", "Harry Bosch").Return(output, nil).Once()
			},
			want: Series{ID: "series-id-1", Title: "Harry Bosch", Books: []BooksOrder{{Order: 1, Book: books.Book{ID: "book-id-1"}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)

			r := NewRepository(mockDynamoDBClient, "series-table")

			got, err := r.GetByTitle(ctx, "Harry Bosch")

			assert.Equal(t, tt.want, got)
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

func (m *MockDynamoDBClient) GetByUniqueKey(ctx context.Context, tableName string, value string) (map[string]types.AttributeValue, error) {
	args := m.Called(ctx, tableName, value)
	return args.Get(0).(map[string]types.AttributeValue), args.Error(1)
}
