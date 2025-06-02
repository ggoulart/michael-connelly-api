package books

import "context"

type StorageBook interface {
	Save(ctx context.Context, book Book) (Book, error)
	GetById(ctx context.Context, bookID string) (Book, error)
	GetByTitle(ctx context.Context, bookTitle string) (Book, error)
}

type Service struct {
	storageBook StorageBook
}

func NewService(storageBook StorageBook) *Service {
	return &Service{storageBook: storageBook}
}

func (s *Service) Create(ctx context.Context, book Book) (Book, error) {
	savedBook, err := s.storageBook.Save(ctx, book)
	if err != nil {
		return Book{}, err
	}

	return savedBook, nil
}

func (s *Service) GetById(ctx context.Context, bookID string) (Book, error) {
	book, err := s.storageBook.GetById(ctx, bookID)
	if err != nil {
		return Book{}, err
	}

	return book, nil
}

func (s *Service) GetByTitle(ctx context.Context, bookTitle string) (Book, error) {
	book, err := s.storageBook.GetByTitle(ctx, bookTitle)
	if err != nil {
		return Book{}, err
	}

	return book, nil
}
