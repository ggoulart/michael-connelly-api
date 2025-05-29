package characters

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ggoulart/michael-connelly-api/internal/books"
	"github.com/google/uuid"
)

var ErrDynamodb = errors.New("dynamodb error")
var ErrNotFound = errors.New("not found")

type DynamoDBClient interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

type Repository struct {
	dynamoDB  DynamoDBClient
	tableName string
	uuidGen   func() uuid.UUID
}

func NewRepository(dynamoDB DynamoDBClient, tableName string, uuidGen func() uuid.UUID) *Repository {
	return &Repository{dynamoDB: dynamoDB, tableName: tableName, uuidGen: uuidGen}
}

func (r *Repository) Save(ctx context.Context, character Character) (Character, error) {
	character.Id = r.uuidGen().String()

	item, err := attributevalue.MarshalMap(NewDBCharacter(character))
	if err != nil {
		return Character{}, fmt.Errorf("%w. failed to marshal character: %w", ErrDynamodb, err)
	}

	_, err = r.dynamoDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return Character{}, fmt.Errorf("%w. failed to save character: %w", ErrDynamodb, err)
	}

	return character, err
}

func (r *Repository) GetById(ctx context.Context, characterID string) (Character, error) {
	output, err := r.dynamoDB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key:       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: characterID}},
	})
	if err != nil {
		return Character{}, fmt.Errorf("%w. failed to fetch character, id: %s, err: %w", ErrDynamodb, characterID, err)
	}
	if output.Item == nil {
		return Character{}, fmt.Errorf("%w. character id: %s", ErrNotFound, characterID)
	}

	var character Character
	err = attributevalue.UnmarshalMap(output.Item, &character)
	if err != nil {
		return Character{}, fmt.Errorf("%w. failed to unmarshal character: %w", ErrDynamodb, err)
	}
	return character, nil
}

func (r *Repository) GetByName(ctx context.Context, characterName string) (Character, error) {
	output, err := r.dynamoDB.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.tableName),
		IndexName:                 aws.String("characters_name"),
		KeyConditionExpression:    aws.String("#n = :name"),
		ExpressionAttributeNames:  map[string]string{"#n": "name"},
		ExpressionAttributeValues: map[string]types.AttributeValue{":name": &types.AttributeValueMemberS{Value: characterName}},
		Limit:                     aws.Int32(1),
	})
	if err != nil {
		return Character{}, fmt.Errorf("%w. failed to query character by name: %w", ErrDynamodb, err)
	}

	if len(output.Items) == 0 {
		return Character{}, fmt.Errorf("%w. character name: %s", ErrNotFound, characterName)
	}

	var character Character
	err = attributevalue.UnmarshalMap(output.Items[0], &character)
	if err != nil {
		return Character{}, fmt.Errorf("%w. failed to unmarshal character: %w", ErrDynamodb, err)
	}

	return character, nil
}

func (r *Repository) AddBooks(ctx context.Context, characterID string, booksList []books.Book) (Character, error) {
	var bookIDs []types.AttributeValue
	for _, b := range booksList {
		bookIDs = append(bookIDs, &types.AttributeValueMemberS{Value: b.Id})
	}

	_, err := r.dynamoDB.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:        aws.String(r.tableName),
		Key:              map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: characterID}},
		UpdateExpression: aws.String("SET books = list_append(if_not_exists(books, :empty_list), :new_books)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":new_books":  &types.AttributeValueMemberL{Value: bookIDs},
			":empty_list": &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
		},
		ReturnValues: types.ReturnValueAllNew,
	})
	if err != nil {
		return Character{}, fmt.Errorf("%w. failed to add books to character: %w", ErrDynamodb, err)
	}

	return Character{}, nil
}

type DBCharacter struct {
	ID    string   `dynamodbav:"id"`
	Name  string   `dynamodbav:"name"`
	Books []string `dynamodbav:"books"`
}

func NewDBCharacter(character Character) DBCharacter {
	bookIds := []string{}
	for _, b := range character.Books {
		bookIds = append(bookIds, b.Id)
	}

	return DBCharacter{
		ID:    character.Id,
		Name:  character.Name,
		Books: bookIds,
	}
}
