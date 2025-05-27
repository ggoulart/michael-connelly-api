package characters

import "github.com/ggoulart/michael-connelly-api/internal/books"

type Character struct {
	Id    string
	Name  string
	Books []books.Book
}
