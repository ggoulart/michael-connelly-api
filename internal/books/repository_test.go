package books

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ggoulart/michael-connelly-api/internal/dynamo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRepository_Save(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(*MockDynamoDBClient)
		want    Book
		wantErr error
	}{
		{
			name: "when failed to save book because already exists",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{
					"id":    &types.AttributeValueMemberS{Value: ""},
					"title": &types.AttributeValueMemberS{Value: "The Black Echo"},
					"blurb": &types.AttributeValueMemberS{Value: "For LAPD homicide cop Harry Bosch — hero, maverick, nighthawk — the body in the drainpipe at Mulholland dam is more than another anonymous statistic.  This one is personal. The dead man, Billy Meadows, was a fellow Vietnam “tunnel rat” who fought side by side with him in a nightmare underground war that brought them to the depths of hell.  Now, Bosch is about to relive the horrors of Nam.  From a dangerous maze of blind alleys to a daring criminal heist beneath the city to the tortuous link that must be uncovered, his survival instincts will once again be tested to their limit. Joining with an enigmatic female FBI agent, pitted against enemies within his own department, Bosch must make the agonizing choice between justice and vengeance, as he tracks down a killer whose true face will shock him. The Black Echo won the Edgar Award for Best First Mystery Novel awarded by the Mystery Writers of America."},
					"year":  &types.AttributeValueMemberN{Value: "1992"},
					"adaptations": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
							"description": &types.AttributeValueMemberS{Value: "Bosch S03"},
							"imdb":        &types.AttributeValueMemberS{Value: "https://www.imdb.com/title/tt3502248/episodes/?season=3"},
						}}}},
				}
				m.On("Save", ctx, "table-name", item, "The Black Echo").Return("", dynamo.ErrDuplicated).Once()
				output := map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "random-id"}, "title": &types.AttributeValueMemberS{Value: "The Black Echo"}}
				m.On("GetByUniqueKey", ctx, "table-name", "The Black Echo").Return(output, nil).Once()
			},
			want: Book{ID: "random-id", Title: "The Black Echo"},
		},
		{
			name: "when failed to save book",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{
					"id":    &types.AttributeValueMemberS{Value: ""},
					"title": &types.AttributeValueMemberS{Value: "The Black Echo"},
					"blurb": &types.AttributeValueMemberS{Value: "For LAPD homicide cop Harry Bosch — hero, maverick, nighthawk — the body in the drainpipe at Mulholland dam is more than another anonymous statistic.  This one is personal. The dead man, Billy Meadows, was a fellow Vietnam “tunnel rat” who fought side by side with him in a nightmare underground war that brought them to the depths of hell.  Now, Bosch is about to relive the horrors of Nam.  From a dangerous maze of blind alleys to a daring criminal heist beneath the city to the tortuous link that must be uncovered, his survival instincts will once again be tested to their limit. Joining with an enigmatic female FBI agent, pitted against enemies within his own department, Bosch must make the agonizing choice between justice and vengeance, as he tracks down a killer whose true face will shock him. The Black Echo won the Edgar Award for Best First Mystery Novel awarded by the Mystery Writers of America."},
					"year":  &types.AttributeValueMemberN{Value: "1992"},
					"adaptations": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
							"description": &types.AttributeValueMemberS{Value: "Bosch S03"},
							"imdb":        &types.AttributeValueMemberS{Value: "https://www.imdb.com/title/tt3502248/episodes/?season=3"},
						}}}},
				}
				m.On("Save", ctx, "table-name", item, "The Black Echo").Return("", assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "when successfully saved book",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{
					"id":    &types.AttributeValueMemberS{Value: ""},
					"title": &types.AttributeValueMemberS{Value: "The Black Echo"},
					"blurb": &types.AttributeValueMemberS{Value: "For LAPD homicide cop Harry Bosch — hero, maverick, nighthawk — the body in the drainpipe at Mulholland dam is more than another anonymous statistic.  This one is personal. The dead man, Billy Meadows, was a fellow Vietnam “tunnel rat” who fought side by side with him in a nightmare underground war that brought them to the depths of hell.  Now, Bosch is about to relive the horrors of Nam.  From a dangerous maze of blind alleys to a daring criminal heist beneath the city to the tortuous link that must be uncovered, his survival instincts will once again be tested to their limit. Joining with an enigmatic female FBI agent, pitted against enemies within his own department, Bosch must make the agonizing choice between justice and vengeance, as he tracks down a killer whose true face will shock him. The Black Echo won the Edgar Award for Best First Mystery Novel awarded by the Mystery Writers of America."},
					"year":  &types.AttributeValueMemberN{Value: "1992"},
					"adaptations": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
							"description": &types.AttributeValueMemberS{Value: "Bosch S03"},
							"imdb":        &types.AttributeValueMemberS{Value: "https://www.imdb.com/title/tt3502248/episodes/?season=3"},
						}}}},
				}
				m.On("Save", ctx, "table-name", item, "The Black Echo").Return("random-id", nil).Once()
			},
			want: Book{ID: "random-id", Title: "The Black Echo", Year: 1992, Blurb: "For LAPD homicide cop Harry Bosch — hero, maverick, nighthawk — the body in the drainpipe at Mulholland dam is more than another anonymous statistic.  This one is personal. The dead man, Billy Meadows, was a fellow Vietnam “tunnel rat” who fought side by side with him in a nightmare underground war that brought them to the depths of hell.  Now, Bosch is about to relive the horrors of Nam.  From a dangerous maze of blind alleys to a daring criminal heist beneath the city to the tortuous link that must be uncovered, his survival instincts will once again be tested to their limit. Joining with an enigmatic female FBI agent, pitted against enemies within his own department, Bosch must make the agonizing choice between justice and vengeance, as he tracks down a killer whose true face will shock him. The Black Echo won the Edgar Award for Best First Mystery Novel awarded by the Mystery Writers of America.", Adaptations: []Adaptation{{Description: "Bosch S03", IMDB: "https://www.imdb.com/title/tt3502248/episodes/?season=3"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)

			r := NewRepository(mockDynamoDBClient, "table-name")

			got, err := r.Save(ctx, Book{
				Title:       "The Black Echo",
				Year:        1992,
				Blurb:       "For LAPD homicide cop Harry Bosch — hero, maverick, nighthawk — the body in the drainpipe at Mulholland dam is more than another anonymous statistic.  This one is personal. The dead man, Billy Meadows, was a fellow Vietnam “tunnel rat” who fought side by side with him in a nightmare underground war that brought them to the depths of hell.  Now, Bosch is about to relive the horrors of Nam.  From a dangerous maze of blind alleys to a daring criminal heist beneath the city to the tortuous link that must be uncovered, his survival instincts will once again be tested to their limit. Joining with an enigmatic female FBI agent, pitted against enemies within his own department, Bosch must make the agonizing choice between justice and vengeance, as he tracks down a killer whose true face will shock him. The Black Echo won the Edgar Award for Best First Mystery Novel awarded by the Mystery Writers of America.",
				Adaptations: []Adaptation{{Description: "Bosch S03", IMDB: "https://www.imdb.com/title/tt3502248/episodes/?season=3"}},
			})

			assert.Equal(t, tt.want, got)
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
			name: "when failed to get book by id",
			setup: func(m *MockDynamoDBClient) {
				m.On("GetByID", ctx, "table-name", "random-id").Return(map[string]types.AttributeValue{}, assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "when failed to unmarshal book",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{"title": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}}}
				m.On("GetByID", ctx, "table-name", "random-id").Return(item, nil).Once()
			},
			wantErr: fmt.Errorf("failed to unmarshal book: %w", &attributevalue.UnmarshalTypeError{Value: "map", Type: reflect.TypeOf("string")}),
		},
		{
			name: "when successfully get book by id",
			setup: func(m *MockDynamoDBClient) {
				item := map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "random-id"}, "title": &types.AttributeValueMemberS{Value: "The Black Echo"}}
				m.On("GetByID", ctx, "table-name", "random-id").Return(item, nil).Once()
			},
			want: Book{ID: "random-id", Title: "The Black Echo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)

			r := NewRepository(mockDynamoDBClient, "table-name")

			got, err := r.GetById(ctx, "random-id")

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
		want    []Book
		wantErr error
	}{
		{
			name: "when failed to get book by title",
			setup: func(m *MockDynamoDBClient) {
				m.On("GetByUniqueKey", ctx, "table-name", "The Black Echo").Return(map[string]types.AttributeValue{}, assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "when failed to unmarshal book",
			setup: func(m *MockDynamoDBClient) {
				output := map[string]types.AttributeValue{"title": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}}}
				m.On("GetByUniqueKey", ctx, "table-name", "The Black Echo").Return(output, nil).Once()
			},
			wantErr: fmt.Errorf("failed to unmarshal book: %w", &attributevalue.UnmarshalTypeError{Value: "map", Type: reflect.TypeOf("string")}),
		},
		{
			name: "when successfully get book by title",
			setup: func(m *MockDynamoDBClient) {
				output := map[string]types.AttributeValue{"title": &types.AttributeValueMemberS{Value: "The Black Echo"}}
				m.On("GetByUniqueKey", ctx, "table-name", "The Black Echo").Return(output, nil).Once()
			},
			want: []Book{{Title: "The Black Echo"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamoDBClient := new(MockDynamoDBClient)
			tt.setup(mockDynamoDBClient)

			r := NewRepository(mockDynamoDBClient, "table-name")
			got, err := r.GetBookListByTitles(ctx, []string{"The Black Echo"})

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
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
