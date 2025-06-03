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
	var createRequest CreateRequest
	if err := ctx.BindJSON(&createRequest); err != nil {
		ctx.Error(err)
		return
	}

	createdBook, err := c.manager.Create(ctx, createRequest.ToBook())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, NewBookResponse(createdBook))
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

	ctx.JSON(http.StatusOK, NewBookResponse(book))
}

type CreateRequest struct {
	Title string `json:"title" binding:"required"`
	Year  int    `json:"year" binding:"required,gte=1956"`
	Blurb string `json:"blurb"`
}

func (r *CreateRequest) ToBook() Book {
	return Book{
		Title: r.Title,
		Year:  r.Year,
		Blurb: r.Blurb,
	}
}

type GetByIDRequest struct {
	BookID string `uri:"bookID" binding:"required"`
}

type BookResponse struct {
	ID    string `json:"id,omitempty"`
	Title string `json:"title" binding:"required"`
	Year  int    `json:"year" binding:"required,gte=1956"`
	Blurb string `json:"blurb"`
}

func NewBookResponse(book Book) BookResponse {
	return BookResponse{
		ID:    book.ID,
		Title: book.Title,
		Year:  book.Year,
		Blurb: book.Blurb,
	}
}
