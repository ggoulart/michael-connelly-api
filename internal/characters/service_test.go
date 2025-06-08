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
	tests := []struct {
		name    string
		setup   func(*StorageCharacterMock, *StorageBookMock)
		want    Character
		wantErr error
	}{
		{
			name: "when failed to get book by title",
			setup: func(_ *StorageCharacterMock, b *StorageBookMock) {
				b.On("GetByTitle", ctx, "The Black Echo").Return(books.Book{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "failed to save character",
			setup: func(c *StorageCharacterMock, b *StorageBookMock) {
				book := books.Book{ID: "random-book-id", Title: "The Black Echo"}
				b.On("GetByTitle", ctx, "The Black Echo").Return(book, nil)
				c.On("Save", ctx, Character{Name: "Harry Bosch", Books: []books.Book{book}}).Return(Character{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "successfully saved character",
			setup: func(c *StorageCharacterMock, b *StorageBookMock) {
				book := books.Book{ID: "random-book-id", Title: "The Black Echo"}
				b.On("GetByTitle", ctx, "The Black Echo").Return(book, nil)
				savedCharacter := Character{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"}
				c.On("Save", ctx, Character{Name: "Harry Bosch", Books: []books.Book{book}}).Return(savedCharacter, nil)
			},
			want: Character{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageCharacter := new(StorageCharacterMock)
			storageBook := new(StorageBookMock)
			tt.setup(storageCharacter, storageBook)

			s := NewService(storageCharacter, storageBook)

			got, err := s.Create(ctx, Character{Name: "Harry Bosch"}, []string{"The Black Echo"})

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
			storageCharacter.AssertExpectations(t)
			storageBook.AssertExpectations(t)
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
				returnedCharacter := Character{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"}
				s.On("GetById", ctx, "a-random-character-id").Return(returnedCharacter, nil)
			},
			want: Character{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"},
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

func TestService_GetByName(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func(s *StorageCharacterMock)
		want    Character
		wantErr error
	}{
		{
			name: "when failed to get character",
			setup: func(m *StorageCharacterMock) {
				m.On("GetByName", ctx, "Harry Bosch").Return(Character{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "successfully get character",
			setup: func(m *StorageCharacterMock) {
				character := Character{ID: "random-id", Name: "Harry Bosch"}
				m.On("GetByName", ctx, "Harry Bosch").Return(character, nil)
			},
			want: Character{ID: "random-id", Name: "Harry Bosch"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageCharacter := new(StorageCharacterMock)
			tt.setup(storageCharacter)

			s := NewService(storageCharacter, nil)
			got, err := s.GetByName(ctx, "Harry Bosch")

			assert.Equal(t, got, tt.want)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

type StorageCharacterMock struct {
	mock.Mock
}

func (s *StorageCharacterMock) GetByName(ctx context.Context, characterName string) (Character, error) {
	args := s.Called(ctx, characterName)
	return args.Get(0).(Character), args.Error(1)
}

func (s *StorageCharacterMock) Save(ctx context.Context, character Character) (Character, error) {
	args := s.Called(ctx, character)
	return args.Get(0).(Character), args.Error(1)
}

func (s *StorageCharacterMock) GetById(ctx context.Context, characterID string) (Character, error) {
	args := s.Called(ctx, characterID)
	return args.Get(0).(Character), args.Error(1)
}

type StorageBookMock struct {
	mock.Mock
}

func (s *StorageBookMock) GetByTitle(ctx context.Context, bookTitle string) (books.Book, error) {
	args := s.Called(ctx, bookTitle)
	return args.Get(0).(books.Book), args.Error(1)
}
