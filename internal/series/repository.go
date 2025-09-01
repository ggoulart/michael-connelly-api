package series

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ggoulart/michael-connelly-api/internal/books"
	"github.com/ggoulart/michael-connelly-api/internal/dynamo"
)

type DynamoDBClient interface {
	Save(ctx context.Context, tableName string, item map[string]types.AttributeValue, uniqueKey string) (string, error)
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

func (r *Repository) Save(ctx context.Context, series Series) (Series, error) {
	seriesItem, err := attributevalue.MarshalMap(NewDBSeries(series))
	if err != nil {
		return Series{}, fmt.Errorf("failed to marshal series: %w", err)
	}

	id, err := r.dynamoDBClient.Save(ctx, r.tableName, seriesItem, series.Title)
	if err != nil {
		if errors.Is(err, dynamo.ErrDuplicated) {
			return r.GetByTitle(ctx, series.Title)
		}
		return Series{}, err
	}

	series.ID = id

	return series, nil
}

func (r *Repository) GetByTitle(ctx context.Context, title string) (Series, error) {
	item, err := r.dynamoDBClient.GetByUniqueKey(ctx, r.tableName, title)
	if err != nil {
		return Series{}, err
	}

	var dbSeries DBSeries
	err = attributevalue.UnmarshalMap(item, &dbSeries)
	if err != nil {
		return Series{}, fmt.Errorf("failed to unmarshal series: %w", err)
	}

	return dbSeries.ToSeries(), nil
}

func (r *Repository) GetAll(ctx context.Context) ([]Series, error) {
	items, err := r.dynamoDBClient.GetAll(ctx, r.tableName)
	if err != nil {
		return []Series{}, err
	}

	var seriesList []Series
	for _, item := range items {
		var dbSeries DBSeries
		err = attributevalue.UnmarshalMap(item, &dbSeries)
		if err != nil {
			return []Series{}, fmt.Errorf("failed to unmarshal series: %w", err)
		}

		seriesList = append(seriesList, dbSeries.ToSeries())
	}

	return seriesList, nil
}

type DBSeries struct {
	ID         string         `dynamodbav:"id"`
	Title      string         `dynamodbav:"title"`
	BooksOrder []DBBooksOrder `dynamodbav:"booksOrder"`
}

type DBBooksOrder struct {
	BookID string `dynamodbav:"book_id"`
	Order  int    `dynamodbav:"order"`
}

func NewDBSeries(series Series) DBSeries {
	booksOrderList := []DBBooksOrder{}

	for _, book := range series.Books {
		booksOrderList = append(booksOrderList, DBBooksOrder{
			BookID: book.Book.ID,
			Order:  book.Order,
		})
	}

	return DBSeries{
		ID:         series.ID,
		Title:      series.Title,
		BooksOrder: booksOrderList,
	}
}

func (d *DBSeries) ToSeries() Series {
	var booksList []BooksOrder

	for _, book := range d.BooksOrder {
		booksList = append(booksList, BooksOrder{
			Book:  books.Book{ID: book.BookID},
			Order: book.Order,
		})
	}

	return Series{
		ID:    d.ID,
		Title: d.Title,
		Books: booksList,
	}
}
