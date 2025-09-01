package series

import (
	"context"

	"github.com/ggoulart/michael-connelly-api/internal/books"
)

type StorageSeries interface {
	Save(ctx context.Context, series Series) (Series, error)
	GetByTitle(ctx context.Context, title string) (Series, error)
	GetAll(ctx context.Context) ([]Series, error)
}

type StorageBook interface {
	GetById(ctx context.Context, bookID string) (books.Book, error)
	GetByTitle(ctx context.Context, bookTitle string) (books.Book, error)
}

type Service struct {
	storageSeries StorageSeries
	storageBook   StorageBook
}

func NewService(storageSeries StorageSeries, storageBook StorageBook) *Service {
	return &Service{storageSeries: storageSeries, storageBook: storageBook}
}

func (s *Service) Create(ctx context.Context, series Series, booksOrderList []BooksOrder) (Series, error) {
	for _, bookOrder := range booksOrderList {
		book, err := s.storageBook.GetByTitle(ctx, bookOrder.Book.Title)
		if err != nil {
			return Series{}, err
		}

		series.Books = append(series.Books, BooksOrder{
			Order: bookOrder.Order,
			Book:  book,
		})
	}

	savedSeries, err := s.storageSeries.Save(ctx, series)
	if err != nil {
		return Series{}, err
	}

	return savedSeries, nil
}

func (s *Service) GetAll(ctx context.Context) ([]Series, error) {
	seriesList, err := s.storageSeries.GetAll(ctx)
	if err != nil {
		return []Series{}, err
	}

	for _, series := range seriesList {
		for i := range series.Books {
			book, err := s.storageBook.GetById(ctx, series.Books[i].ID)
			if err != nil {
				return []Series{}, err
			}

			series.Books[i].Book = book
		}
	}

	return seriesList, nil
}
