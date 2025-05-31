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
	var bookRequest BookRequest
	if err := ctx.BindJSON(&bookRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	createdBook, err := c.manager.Create(ctx, bookRequest.ToBook())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid error"})
		return
	}

	ctx.JSON(http.StatusCreated, NewBookResponse(createdBook))
}

func (c *Controller) GetById(ctx *gin.Context) {
	bookID := ctx.Param("bookID")
	if bookID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bookID is required"})
		return
	}

	book, err := c.manager.GetById(ctx, bookID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, NewBookResponse(book))
}

type BookRequest struct {
	Title string `json:"title" binding:"required"`
	Year  int    `json:"year" binding:"required,gte=1956"`
	Blurb string `json:"blurb"`
}

func (r *BookRequest) ToBook() Book {
	return Book{
		Title: r.Title,
		Year:  r.Year,
		Blurb: r.Blurb,
	}
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
