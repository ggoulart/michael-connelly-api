package books

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Manager interface {
	Create(ctx context.Context, book Book) (Book, error)
	GetById(ctx context.Context, bookID string) (Book, error)
}

type Controller struct {
	manager Manager
}

func NewController(manager Manager) *Controller {
	return &Controller{manager: manager}
}

func (c *Controller) Create(ctx *gin.Context) {
	var bookDTO BookDTO
	if err := ctx.BindJSON(&bookDTO); err != nil {
		ctx.Error(err)
		return
	}

	createdBook, err := c.manager.Create(ctx, bookDTO.ToBook())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, NewBookDTO(createdBook))
}

func (c *Controller) GetById(ctx *gin.Context) {
	var getByIDRequest GetByIDRequest
	if err := ctx.BindUri(&getByIDRequest); err != nil {
		ctx.Error(err)
		return
	}

	book, err := c.manager.GetById(ctx, getByIDRequest.BookID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, NewBookDTO(book))
}

type BookDTO struct {
	ID          string          `json:"id,omitempty"`
	Title       string          `json:"title" binding:"required"`
	Year        int             `json:"year" binding:"required,gte=1956"`
	Blurb       string          `json:"blurb"`
	Adaptations []AdaptationDTO `json:"adaptations"`
}

type AdaptationDTO struct {
	Description string `json:"description" binding:"required"`
	IMDB        string `json:"imdb" binding:"required"`
}

func NewBookDTO(book Book) BookDTO {
	var adaptations []AdaptationDTO
	for _, a := range book.Adaptations {
		adaptations = append(adaptations, AdaptationDTO{
			Description: a.Description,
			IMDB:        a.IMDB,
		})
	}

	return BookDTO{
		ID:          book.ID,
		Title:       book.Title,
		Year:        book.Year,
		Blurb:       book.Blurb,
		Adaptations: adaptations,
	}
}

func (r *BookDTO) ToBook() Book {
	var adaptations []Adaptation
	for _, a := range r.Adaptations {
		adaptations = append(adaptations, Adaptation{
			Description: a.Description,
			IMDB:        a.IMDB,
		})
	}

	return Book{
		Title:       r.Title,
		Year:        r.Year,
		Blurb:       r.Blurb,
		Adaptations: adaptations,
	}
}

type GetByIDRequest struct {
	BookID string `uri:"bookID" binding:"required"`
}
