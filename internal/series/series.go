package series

import "github.com/ggoulart/michael-connelly-api/internal/books"

type Series struct {
	ID    string
	Title string
	Books []BooksOrder
}

type BooksOrder struct {
	Order int
	books.Book
}
