package books

type Book struct {
	Id    string `json:"id,omitempty"`
	Title string `json:"title" binding:"required"`
	Year  int    `json:"year" binding:"required,gte=1956"`
	Blurb string `json:"blurb"`
}
