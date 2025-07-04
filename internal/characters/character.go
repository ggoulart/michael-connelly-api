package characters

import (
	"github.com/ggoulart/michael-connelly-api/internal/books"
)

type Character struct {
	ID     string
	Name   string
	Books  []books.Book
	Actors []Actor
}

type Actor struct {
	Name string
	IMDB string
}
