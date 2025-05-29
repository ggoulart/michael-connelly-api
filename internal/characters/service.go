package characters

import (
	"context"

	"github.com/ggoulart/michael-connelly-api/internal/books"
)

type StorageCharacter interface {
	Save(ctx context.Context, character Character) (Character, error)
	GetById(ctx context.Context, characterID string) (Character, error)
	AddBooks(ctx context.Context, characterID string, booksList []books.Book) (Character, error)
}

type StorageBooks interface {
	GetByNames(ctx context.Context, booksTitles []string) ([]books.Book, error)
}

type Service struct {
	storageCharacter StorageCharacter
	storageBooks     StorageBooks
}

func NewService(storageCharacter StorageCharacter, storageBooks StorageBooks) *Service {
	return &Service{storageCharacter: storageCharacter, storageBooks: storageBooks}
}

func (s *Service) Create(ctx context.Context, character Character) (Character, error) {
	savedCharacter, err := s.storageCharacter.Save(ctx, character)
	if err != nil {
		return Character{}, err
	}

	return savedCharacter, nil
}

func (s *Service) GetById(ctx context.Context, characterID string) (Character, error) {
	character, err := s.storageCharacter.GetById(ctx, characterID)
	if err != nil {
		return Character{}, err
	}

	return character, nil
}

func (s *Service) AddBooks(ctx context.Context, characterID string, booksNames []string) (Character, error) {
	booksList, err := s.storageBooks.GetByNames(ctx, booksNames)
	if err != nil {
		return Character{}, err
	}

	character, err := s.storageCharacter.AddBooks(ctx, characterID, booksList)
	if err != nil {
		return Character{}, err
	}

	return character, nil
}
