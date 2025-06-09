package characters

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ggoulart/michael-connelly-api/internal/dynamo"
)

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
		return Character{}, fmt.Errorf("failed to marshal character: %w", err)
	}

	id, err := r.dynamodb.Save(ctx, r.tableName, characterItem, character.Name)
	if err != nil {
		if errors.Is(err, dynamo.ErrDuplicated) {
			return r.GetByName(ctx, character.Name)
		}
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

	var dbCharacter DBCharacter
	err = attributevalue.UnmarshalMap(item, &dbCharacter)
	if err != nil {
		return Character{}, fmt.Errorf("failed to unmarshal character: %w", err)
	}

	return dbCharacter.ToCharacter(), nil
}

func (r *Repository) GetByName(ctx context.Context, characterName string) (Character, error) {
	item, err := r.dynamodb.GetByUniqueKey(ctx, r.tableName, characterName)
	if err != nil {
		return Character{}, err
	}

	var dbCharacter DBCharacter
	err = attributevalue.UnmarshalMap(item, &dbCharacter)
	if err != nil {
		return Character{}, fmt.Errorf("failed to unmarshal character: %w", err)
	}

	return dbCharacter.ToCharacter(), nil
}

type DBCharacter struct {
	ID     string    `dynamodbav:"id"`
	Name   string    `dynamodbav:"name"`
	Books  []string  `dynamodbav:"books"`
	Actors []DBActor `dynamodbav:"actors"`
}

type DBActor struct {
	Name string `dynamodbav:"name"`
	IMDB string `dynamodbav:"imdb"`
}

func NewDBCharacter(character Character) DBCharacter {
	bookIds := []string{}
	for _, b := range character.Books {
		bookIds = append(bookIds, b.ID)
	}

	var actors []DBActor
	for _, a := range character.Actors {
		actors = append(actors, DBActor{Name: a.Name, IMDB: a.IMDB})
	}

	return DBCharacter{
		ID:     character.ID,
		Name:   character.Name,
		Books:  bookIds,
		Actors: actors,
	}
}

func (d *DBCharacter) ToCharacter() Character {
	return Character{
		ID:   d.ID,
		Name: d.Name,
	}
}
