package books

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

var ErrDynamodb = errors.New("dynamodb error")
var ErrNotFound = errors.New("not found")

type DynamoDBClient interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
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

func (r *Repository) Save(ctx context.Context, book Book) (Book, error) {
	book.Id = r.uuidGen().String()

	item, err := attributevalue.MarshalMap(NewDBBook(book))
	if err != nil {
		return Book{}, fmt.Errorf("%w. failed to marshal book: %w", ErrDynamodb, err)
	}

	_, err = r.dynamoDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return Book{}, fmt.Errorf("%w. failed to save book: %w", ErrDynamodb, err)
	}

	return book, err
}

func (r *Repository) GetById(ctx context.Context, bookID string) (Book, error) {
	output, err := r.dynamoDB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key:       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: bookID}},
	})
	if err != nil {
		return Book{}, fmt.Errorf("%w. failed to fetch book, id: %s, err: %w", ErrDynamodb, bookID, err)
	}
	if output.Item == nil {
		return Book{}, fmt.Errorf("%w. book id: %s", ErrNotFound, bookID)
	}

	var book Book
	err = attributevalue.UnmarshalMap(output.Item, &book)
	if err != nil {
		return Book{}, fmt.Errorf("%w. failed to unmarshal book: %w", ErrDynamodb, err)
	}

	return book, nil
}

func (r *Repository) GetByNames(ctx context.Context, bookTitles []string) ([]Book, error) {
	var books []Book

	for _, title := range bookTitles {
		output, err := r.dynamoDB.Query(ctx, &dynamodb.QueryInput{
			TableName:                 aws.String(r.tableName),
			IndexName:                 aws.String("books_title"),
			KeyConditionExpression:    aws.String("title = :title"),
			ExpressionAttributeValues: map[string]types.AttributeValue{":title": &types.AttributeValueMemberS{Value: title}},
			Limit:                     aws.Int32(1),
		})
		if err != nil {
			return nil, fmt.Errorf("%w. failed to fetch book, title: %s, err: %w", ErrDynamodb, title, err)
		}

		var dbBook DBBook
		err = attributevalue.UnmarshalMap(output.Items[0], &dbBook)
		if err != nil {
			return nil, fmt.Errorf("%w. failed to unmarshal book: %w", ErrDynamodb, err)
		}

		books = append(books, dbBook.ToBook())
	}

	return books, nil
}

type DBBook struct {
	ID    string `dynamodbav:"id"`
	Title string `dynamodbav:"title"`
	Year  int    `dynamodbav:"year"`
	Blurb string `dynamodbav:"blurb"`
}

func NewDBBook(book Book) DBBook {
	return DBBook{
		ID:    book.Id,
		Title: book.Title,
		Year:  book.Year,
		Blurb: book.Blurb,
	}
}

func (b *DBBook) ToBook() Book {
	return Book{
		Id:    b.ID,
		Title: b.Title,
		Year:  b.Year,
		Blurb: b.Blurb,
	}
}
