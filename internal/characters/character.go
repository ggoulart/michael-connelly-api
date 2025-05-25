package characters

type Character struct {
	Id    string   `json:"id,omitempty"`
	Name  string   `json:"name" validate:"required"`
	Books []string `json:"books,omitempty"`
}
