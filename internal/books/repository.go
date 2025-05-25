package books

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

func (r *Repository) Save(ctx context.Context, book Book) (Book, error) {
	book.Id = r.uuidGen().String()

	item, err := attributevalue.MarshalMap(book)
	if err != nil {
		return Book{}, fmt.Errorf("failed to marshal book: %w", err)
	}

	_, err = r.dynamoDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return Book{}, fmt.Errorf("failed to save book: %w", err)
	}

	return book, err
}

func (r *Repository) GetById(ctx context.Context, bookID string) (Book, error) {
	output, err := r.dynamoDB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key:       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: bookID}},
	})
	if err != nil {
		return Book{}, fmt.Errorf("failed to fetch book, id: %s, err: %w", bookID, err)
	}
	if output.Item == nil {
		return Book{}, fmt.Errorf("book not found id: %s", bookID)
	}

	var book Book
	err = attributevalue.UnmarshalMap(output.Item, &book)
	if err != nil {
		return Book{}, fmt.Errorf("failed to unmarshal book: %w", err)
	}
	return book, nil
}
