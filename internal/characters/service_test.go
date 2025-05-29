package characters

import (
	"context"
	"testing"

	"github.com/ggoulart/michael-connelly-api/internal/books"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	receivedCharacter := Character{Name: "Harry Bosch"}
	tests := []struct {
		name    string
		setup   func(s *StorageCharacterMock)
		want    Character
		wantErr error
	}{
		{
			name: "failed to save character",
			setup: func(s *StorageCharacterMock) {
				s.On("Save", ctx, receivedCharacter).Return(Character{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "successfully saved character",
			setup: func(s *StorageCharacterMock) {
				savedCharacter := Character{Id: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"}
				s.On("Save", ctx, receivedCharacter).Return(savedCharacter, nil)
			},
			want: Character{Id: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageCharacter := new(StorageCharacterMock)
			tt.setup(storageCharacter)

			s := NewService(storageCharacter, nil)

			got, err := s.Create(ctx, receivedCharacter)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestService_GetById(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(s *StorageCharacterMock)
		want    Character
		wantErr error
	}{
		{
			name: "failed to get character",
			setup: func(s *StorageCharacterMock) {
				s.On("GetById", ctx, "a-random-character-id").Return(Character{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "successfully saved character",
			setup: func(s *StorageCharacterMock) {
				returnedCharacter := Character{Id: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"}
				s.On("GetById", ctx, "a-random-character-id").Return(returnedCharacter, nil)
			},
			want: Character{Id: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageCharacter := new(StorageCharacterMock)
			tt.setup(storageCharacter)

			s := NewService(storageCharacter, nil)

			got, err := s.GetById(ctx, "a-random-character-id")

			assert.Equal(t, got, tt.want)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestService_AddBooks(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(c *StorageCharacterMock, b *StorageBooksMock)
		want    Character
		wantErr error
	}{
		{
			name: "when failed to GetByNames",
			setup: func(s *StorageCharacterMock, b *StorageBooksMock) {
				b.On("GetByNames", ctx, []string{"The Black Echo", "The Black Ice"}).Return([]books.Book{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "when failed to add books",
			setup: func(s *StorageCharacterMock, b *StorageBooksMock) {
				booksList := []books.Book{{Id: "the-black-echo-id", Title: "The Black Echo"}, {Id: "the-black-ice-id", Title: "The Black Ice"}}
				b.On("GetByNames", ctx, []string{"The Black Echo", "The Black Ice"}).Return(booksList, nil)
				s.On("AddBooks", ctx, "character-id", booksList).Return(Character{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "when successfully add books",
			setup: func(s *StorageCharacterMock, b *StorageBooksMock) {
				booksList := []books.Book{{Id: "the-black-echo-id", Title: "The Black Echo"}, {Id: "the-black-ice-id", Title: "The Black Ice"}}
				b.On("GetByNames", ctx, []string{"The Black Echo", "The Black Ice"}).Return(booksList, nil)
				s.On("AddBooks", ctx, "character-id", booksList).Return(Character{Name: "Harry Bosch"}, nil)
			},
			want: Character{Name: "Harry Bosch"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageCharacter := new(StorageCharacterMock)
			storageBook := new(StorageBooksMock)

			tt.setup(storageCharacter, storageBook)

			s := NewService(storageCharacter, storageBook)

			got, err := s.AddBooks(ctx, "character-id", []string{"The Black Echo", "The Black Ice"})

			assert.Equal(t, got, tt.want)
			assert.Equal(t, err, tt.wantErr)
			storageBook.AssertExpectations(t)
			storageCharacter.AssertExpectations(t)
		})
	}
}

type StorageCharacterMock struct {
	mock.Mock
}

func (s *StorageCharacterMock) Save(ctx context.Context, character Character) (Character, error) {
	args := s.Called(ctx, character)
	return args.Get(0).(Character), args.Error(1)
}

func (s *StorageCharacterMock) GetById(ctx context.Context, characterID string) (Character, error) {
	args := s.Called(ctx, characterID)
	return args.Get(0).(Character), args.Error(1)
}

func (s *StorageCharacterMock) AddBooks(ctx context.Context, characterID string, books []books.Book) (Character, error) {
	args := s.Called(ctx, characterID, books)
	return args.Get(0).(Character), args.Error(1)
}

type StorageBooksMock struct {
	mock.Mock
}

func (s *StorageBooksMock) GetByNames(ctx context.Context, booksTitles []string) ([]books.Book, error) {
	args := s.Called(ctx, booksTitles)
	return args.Get(0).([]books.Book), args.Error(1)
}
