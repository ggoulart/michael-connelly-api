package characters

import (
	"context"

	"github.com/ggoulart/michael-connelly-api/internal/books"
)

type StorageCharacter interface {
	Save(ctx context.Context, character Character) (Character, error)
	GetById(ctx context.Context, characterID string) (Character, error)
	GetByName(ctx context.Context, characterName string) (Character, error)
}

type StorageBook interface {
	GetByTitle(ctx context.Context, bookTitle string) (books.Book, error)
}

type Service struct {
	storageCharacter StorageCharacter
	storageBook      StorageBook
}

func NewService(storageCharacter StorageCharacter, storageBook StorageBook) *Service {
	return &Service{storageCharacter: storageCharacter, storageBook: storageBook}
}

func (s *Service) Create(ctx context.Context, character Character, bookTitles []string) (Character, error) {
	booksList := []books.Book{}

	for _, bookTitle := range bookTitles {
		book, err := s.storageBook.GetByTitle(ctx, bookTitle)
		if err != nil {
			return Character{}, err
		}

		booksList = append(booksList, book)
	}

	character.Books = booksList

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

func (s *Service) GetByName(ctx context.Context, characterName string) (Character, error) {
	character, err := s.storageCharacter.GetByName(ctx, characterName)
	if err != nil {
		return Character{}, err
	}

	return character, nil
}
