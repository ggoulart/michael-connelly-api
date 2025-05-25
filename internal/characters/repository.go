package characters

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type DynamoDBClient interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
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

	item, err := attributevalue.MarshalMap(character)
	if err != nil {
		return Character{}, fmt.Errorf("failed to marshal character: %w", err)
	}

	_, err = r.dynamoDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return Character{}, fmt.Errorf("failed to save character: %w", err)
	}

	return character, err
}

func (r *Repository) GetById(ctx context.Context, characterID string) (Character, error) {
	output, err := r.dynamoDB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key:       map[string]types.AttributeValue{"Id": &types.AttributeValueMemberS{Value: characterID}},
	})
	if err != nil {
		return Character{}, fmt.Errorf("failed to fetch character, id: %s, err: %w", characterID, err)
	}
	if output.Item == nil {
		return Character{}, fmt.Errorf("character not found id: %s", characterID)
	}

	var character Character
	err = attributevalue.UnmarshalMap(output.Item, &character)
	if err != nil {
		return Character{}, fmt.Errorf("failed to unmarshal character: %w", err)
	}
	return character, nil
}
