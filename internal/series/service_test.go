package series

import (
	"context"
	"testing"

	"github.com/ggoulart/michael-connelly-api/internal/books"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(*StorageSeriesMock, *StorageBookMock)
		want    Series
		wantErr error
	}{
		{
			name: "when failed to get book by title",
			setup: func(_ *StorageSeriesMock, b *StorageBookMock) {
				b.On("GetByTitle", ctx, "The Black Echo").Return(books.Book{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "when failed to save series",
			setup: func(s *StorageSeriesMock, b *StorageBookMock) {
				getByTitleOutput := books.Book{Title: "The Black Echo"}
				b.On("GetByTitle", ctx, "The Black Echo").Return(getByTitleOutput, nil)
				s.On("Save", ctx, Series{Title: "Harry Bosch", Books: []BooksOrder{{Order: 1, Book: getByTitleOutput}}}).Return(Series{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "when successful to save series",
			setup: func(s *StorageSeriesMock, b *StorageBookMock) {
				getByTitleOutput := books.Book{ID: "the-black-echo-book-id", Title: "The Black Echo"}
				b.On("GetByTitle", ctx, "The Black Echo").Return(getByTitleOutput, nil)
				saveInput := Series{Title: "Harry Bosch", Books: []BooksOrder{{Order: 1, Book: getByTitleOutput}}}
				savedSeries := Series{ID: "harry-bosch-series-id", Title: "Harry Bosch", Books: []BooksOrder{{Order: 1, Book: getByTitleOutput}}}
				s.On("Save", ctx, saveInput).Return(savedSeries, nil)
			},
			want: Series{ID: "harry-bosch-series-id", Title: "Harry Bosch", Books: []BooksOrder{{Order: 1, Book: books.Book{ID: "the-black-echo-book-id", Title: "The Black Echo"}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageSeries := new(StorageSeriesMock)
			storageBook := new(StorageBookMock)
			tt.setup(storageSeries, storageBook)

			s := NewService(storageSeries, storageBook)

			got, err := s.Create(ctx, Series{Title: "Harry Bosch"}, []BooksOrder{{Order: 1, Book: books.Book{Title: "The Black Echo"}}})

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
			storageSeries.AssertExpectations(t)
			storageBook.AssertExpectations(t)
		})
	}
}

type StorageSeriesMock struct {
	mock.Mock
}

func (s *StorageSeriesMock) Save(ctx context.Context, series Series) (Series, error) {
	args := s.Called(ctx, series)
	return args.Get(0).(Series), args.Error(1)
}

func (s *StorageSeriesMock) GetByTitle(ctx context.Context, title string) (Series, error) {
	args := s.Called(ctx, title)
	return args.Get(0).(Series), args.Error(1)
}

type StorageBookMock struct {
	mock.Mock
}

func (s *StorageBookMock) GetByTitle(ctx context.Context, bookTitle string) (books.Book, error) {
	args := s.Called(ctx, bookTitle)
	return args.Get(0).(books.Book), args.Error(1)
}
