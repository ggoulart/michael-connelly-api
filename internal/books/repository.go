package books

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var ErrDynamodb = errors.New("dynamodb error")

type DynamoDBClient interface {
	Save(ctx context.Context, tableName string, item map[string]types.AttributeValue, uniqueKey string) (string, error)
	GetByID(ctx context.Context, tableName string, id string) (map[string]types.AttributeValue, error)
	GetByUniqueKey(ctx context.Context, tableName string, value string) (map[string]types.AttributeValue, error)
}

type Repository struct {
	dynamoDBClient DynamoDBClient
	tableName      string
}

func NewRepository(dynamoDBClient DynamoDBClient, tableName string) *Repository {
	return &Repository{dynamoDBClient: dynamoDBClient, tableName: tableName}
}

func (r *Repository) Save(ctx context.Context, book Book) (Book, error) {
	bookItem, err := attributevalue.MarshalMap(NewDBBook(book))
	if err != nil {
		return Book{}, fmt.Errorf("%w. failed to marshal book: %w", ErrDynamodb, err)
	}

	id, err := r.dynamoDBClient.Save(ctx, r.tableName, bookItem, book.Title)
	if err != nil {
		return Book{}, err
	}

	book.ID = id

	return book, err
}

func (r *Repository) GetById(ctx context.Context, bookID string) (Book, error) {
	item, err := r.dynamoDBClient.GetByID(ctx, r.tableName, bookID)
	if err != nil {
		return Book{}, err
	}

	var dbBook DBBook
	err = attributevalue.UnmarshalMap(item, &dbBook)
	if err != nil {
		return Book{}, fmt.Errorf("%w. failed to unmarshal book: %w", ErrDynamodb, err)
	}

	return dbBook.ToBook(), nil
}

func (r *Repository) GetByNames(ctx context.Context, bookTitles []string) ([]Book, error) {
	var booksList []Book

	for _, bookTitle := range bookTitles {
		item, err := r.dynamoDBClient.GetByUniqueKey(ctx, r.tableName, bookTitle)
		if err != nil {
			return nil, err
		}

		var dbBook DBBook
		err = attributevalue.UnmarshalMap(item, &dbBook)
		if err != nil {
			return nil, fmt.Errorf("%w. failed to unmarshal book: %w", ErrDynamodb, err)
		}

		booksList = append(booksList, dbBook.ToBook())
	}

	return booksList, nil
}

type DBBook struct {
	ID    string `dynamodbav:"id"`
	Title string `dynamodbav:"title"`
	Year  int    `dynamodbav:"year"`
	Blurb string `dynamodbav:"blurb"`
}

func NewDBBook(book Book) DBBook {
	return DBBook{
		ID:    book.ID,
		Title: book.Title,
		Year:  book.Year,
		Blurb: book.Blurb,
	}
}

func (b *DBBook) ToBook() Book {
	return Book{
		ID:    b.ID,
		Title: b.Title,
		Year:  b.Year,
		Blurb: b.Blurb,
	}
}
