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
	var book Book
	if err := ctx.BindJSON(&book); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	createdBook, err := c.manager.Create(ctx, book)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid error"})
		return
	}

	ctx.JSON(http.StatusCreated, createdBook)
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

	ctx.JSON(http.StatusOK, book)
}
