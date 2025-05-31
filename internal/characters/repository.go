package characters

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var ErrDynamodb = errors.New("dynamodb error")

type DynamoClient interface {
	Save(ctx context.Context, tableName string, item map[string]types.AttributeValue, uniqueKey string) (string, error)
	GetByID(ctx context.Context, tableName string, id string) (map[string]types.AttributeValue, error)
	GetByUniqueKey(ctx context.Context, tableName string, value string) (map[string]types.AttributeValue, error)
}

type Repository struct {
	dynamodb  DynamoClient
	tableName string
}

func NewRepository(dynamoDB DynamoClient, tableName string) *Repository {
	return &Repository{dynamodb: dynamoDB, tableName: tableName}
}

func (r *Repository) Save(ctx context.Context, character Character) (Character, error) {
	characterItem, err := attributevalue.MarshalMap(NewDBCharacter(character))
	if err != nil {
		return Character{}, fmt.Errorf("%w. failed to marshal character: %w", ErrDynamodb, err)
	}

	id, err := r.dynamodb.Save(ctx, r.tableName, characterItem, character.Name)
	if err != nil {
		return Character{}, err
	}

	character.ID = id

	return character, nil
}

func (r *Repository) GetById(ctx context.Context, characterID string) (Character, error) {
	item, err := r.dynamodb.GetByID(ctx, r.tableName, characterID)
	if err != nil {
		return Character{}, err
	}

	var character Character
	err = attributevalue.UnmarshalMap(item, &character)
	if err != nil {
		return Character{}, fmt.Errorf("%w. failed to unmarshal character: %w", ErrDynamodb, err)
	}
	return character, nil
}

func (r *Repository) GetByName(ctx context.Context, characterName string) (Character, error) {
	item, err := r.dynamodb.GetByUniqueKey(ctx, r.tableName, characterName)
	if err != nil {
		return Character{}, err
	}

	var character Character
	err = attributevalue.UnmarshalMap(item, &character)
	if err != nil {
		return Character{}, fmt.Errorf("%w. failed to unmarshal character: %w", ErrDynamodb, err)
	}

	return character, nil
}

type DBCharacter struct {
	ID    string   `dynamodbav:"id"`
	Name  string   `dynamodbav:"name"`
	Books []string `dynamodbav:"books"`
}

func NewDBCharacter(character Character) DBCharacter {
	bookIds := []string{}
	for _, b := range character.Books {
		bookIds = append(bookIds, b.ID)
	}

	return DBCharacter{
		ID:    character.ID,
		Name:  character.Name,
		Books: bookIds,
	}
}
