package books

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ggoulart/michael-connelly-api/internal/dynamo"
)

type DynamoDBClient interface {
	Save(ctx context.Context, tableName string, item map[string]types.AttributeValue, uniqueKey string) (string, error)
	GetByID(ctx context.Context, tableName string, id string) (map[string]types.AttributeValue, error)
	GetByUniqueKey(ctx context.Context, tableName string, value string) (map[string]types.AttributeValue, error)
	GetAll(ctx context.Context, tableName string) ([]map[string]types.AttributeValue, error)
}

type Repository struct {
	dynamoDBClient DynamoDBClient
	tableName      string
}

func NewRepository(dynamoDBClient DynamoDBClient, tableName string) *Repository {
	return &Repository{dynamoDBClient: dynamoDBClient, tableName: tableName}
}

func (r *Repository) Save(ctx context.Context, book Book) (Book, error) {
	bookItem, err := attributevalue.MarshalMap(newDBBook(book))
	if err != nil {
		return Book{}, fmt.Errorf("failed to marshal book: %w", err)
	}

	id, err := r.dynamoDBClient.Save(ctx, r.tableName, bookItem, book.Title)
	if err != nil {
		if errors.Is(err, dynamo.ErrDuplicated) {
			return r.GetByTitle(ctx, book.Title)
		}
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
		return Book{}, fmt.Errorf("failed to unmarshal book: %w", err)
	}

	return dbBook.toBook(), nil
}

func (r *Repository) GetByTitle(ctx context.Context, bookTitle string) (Book, error) {
	item, err := r.dynamoDBClient.GetByUniqueKey(ctx, r.tableName, bookTitle)
	if err != nil {
		return Book{}, err
	}

	var dbBook DBBook
	err = attributevalue.UnmarshalMap(item, &dbBook)
	if err != nil {
		return Book{}, fmt.Errorf("failed to unmarshal book: %w", err)
	}

	return dbBook.toBook(), nil
}

func (r *Repository) GetBookListByTitles(ctx context.Context, bookTitles []string) ([]Book, error) {
	var booksList []Book

	for _, bookTitle := range bookTitles {
		book, err := r.GetByTitle(ctx, bookTitle)
		if err != nil {
			return nil, err
		}

		booksList = append(booksList, book)
	}

	return booksList, nil
}

func (r *Repository) GetAll(ctx context.Context) ([]Book, error) {
	items, err := r.dynamoDBClient.GetAll(ctx, r.tableName)
	if err != nil {
		return []Book{}, err
	}

	var booksList []Book
	for _, item := range items {
		var dbBook DBBook
		err = attributevalue.UnmarshalMap(item, &dbBook)
		if err != nil {
			return []Book{}, fmt.Errorf("failed to unmarshal book: %w", err)
		}

		booksList = append(booksList, dbBook.toBook())
	}

	return booksList, nil
}

type DBBook struct {
	ID          string         `dynamodbav:"id"`
	Title       string         `dynamodbav:"title"`
	Year        int            `dynamodbav:"year"`
	Blurb       string         `dynamodbav:"blurb"`
	Adaptations []DBAdaptation `dynamodbav:"adaptations"`
}

type DBAdaptation struct {
	Description string `dynamodbav:"description"`
	IMDB        string `dynamodbav:"imdb"`
}

func newDBBook(book Book) DBBook {
	var adaptations []DBAdaptation

	for _, a := range book.Adaptations {
		adaptations = append(adaptations, DBAdaptation{
			Description: a.Description,
			IMDB:        a.IMDB,
		})
	}

	return DBBook{
		ID:          book.ID,
		Title:       book.Title,
		Year:        book.Year,
		Blurb:       book.Blurb,
		Adaptations: adaptations,
	}
}

func (b *DBBook) toBook() Book {
	var adaptations []Adaptation

	for _, a := range b.Adaptations {
		adaptations = append(adaptations, Adaptation{
			Description: a.Description,
			IMDB:        a.IMDB,
		})
	}

	return Book{
		ID:          b.ID,
		Title:       b.Title,
		Year:        b.Year,
		Blurb:       b.Blurb,
		Adaptations: adaptations,
	}
}
