package books

type Book struct {
	Id    string `json:"id,omitempty"`
	Title string `json:"title" validate:"required"`
	Year  int    `json:"year" validate:"required,gte=1956"`
	Blurb string `json:"blurb"`
}
