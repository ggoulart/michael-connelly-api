package books

type Book struct {
	ID          string
	Title       string
	Year        int
	Blurb       string
	Adaptations []Adaptation
}

type Adaptation struct {
	Description string
	IMDB        string
}
