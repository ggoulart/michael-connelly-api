package books

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	receivedBook := Book{Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"}
	tests := []struct {
		name    string
		setup   func(s *StorageMock)
		want    Book
		wantErr error
	}{
		{
			name: "failed to save book",
			setup: func(s *StorageMock) {
				s.On("Save", ctx, receivedBook).Return(Book{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "successfully saved book",
			setup: func(s *StorageMock) {
				savedBook := Book{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"}
				s.On("Save", ctx, receivedBook).Return(savedBook, nil)
			},
			want: Book{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := new(StorageMock)
			tt.setup(storage)

			s := NewService(storage)

			got, err := s.Create(context.Background(), receivedBook)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestService_GetById(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(s *StorageMock)
		want    Book
		wantErr error
	}{
		{
			name: "failed to get book",
			setup: func(s *StorageMock) {
				s.On("GetById", ctx, "a-random-book-id").Return(Book{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "successfully saved book",
			setup: func(s *StorageMock) {
				returnedBook := Book{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"}
				s.On("GetById", ctx, "a-random-book-id").Return(returnedBook, nil)
			},
			want: Book{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := new(StorageMock)
			tt.setup(storage)

			s := NewService(storage)

			got, err := s.GetById(context.Background(), "a-random-book-id")

			assert.Equal(t, got, tt.want)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestService_GetByTitle(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(s *StorageMock)
		want    Book
		wantErr error
	}{
		{
			name: "failed to get book",
			setup: func(s *StorageMock) {
				s.On("GetByTitle", ctx, "The Black Echo").Return(Book{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "successfully saved book",
			setup: func(s *StorageMock) {
				returnedBook := Book{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"}
				s.On("GetByTitle", ctx, "The Black Echo").Return(returnedBook, nil)
			},
			want: Book{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := new(StorageMock)
			tt.setup(storage)

			s := NewService(storage)

			got, err := s.GetByTitle(ctx, "The Black Echo")

			assert.Equal(t, got, tt.want)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

type StorageMock struct {
	mock.Mock
}

func (s *StorageMock) Save(ctx context.Context, book Book) (Book, error) {
	args := s.Called(ctx, book)
	return args.Get(0).(Book), args.Error(1)
}

func (s *StorageMock) GetById(ctx context.Context, bookID string) (Book, error) {
	args := s.Called(ctx, bookID)
	return args.Get(0).(Book), args.Error(1)
}

func (s *StorageMock) GetByTitle(ctx context.Context, bookTitle string) (Book, error) {
	args := s.Called(ctx, bookTitle)
	return args.Get(0).(Book), args.Error(1)
}
