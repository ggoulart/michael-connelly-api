package series

import (
	"context"
	"net/http"

	"github.com/ggoulart/michael-connelly-api/internal/books"
	"github.com/gin-gonic/gin"
)

type Manager interface {
	Create(ctx context.Context, series Series, booksOrderList []BooksOrder) (Series, error)
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

	createdSeries, err := c.manager.Create(ctx, createRequest.ToSeries(), createRequest.ToBooksOrderList())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, NewSeriesResponse(createdSeries))
}

type CreateRequest struct {
	Title string `json:"title" binding:"required"`
	Books []struct {
		BookTitle string `json:"bookTitle" binding:"required"`
		Order     int    `json:"order" binding:"required"`
	} `json:"books" binding:"required"`
}

func (r *CreateRequest) ToSeries() Series {
	return Series{
		Title: r.Title,
	}
}

func (r *CreateRequest) ToBooksOrderList() []BooksOrder {
	var booksOrderList []BooksOrder

	for _, b := range r.Books {
		booksOrderList = append(booksOrderList, BooksOrder{
			Order: b.Order,
			Book: books.Book{
				Title: b.BookTitle,
			},
		})
	}

	return booksOrderList
}

type SeriesResponse struct {
	ID    string               `json:"id"`
	Title string               `json:"title"`
	Books []BooksOrderResponse `json:"books"`
}

type BooksOrderResponse struct {
	BookTitle string `json:"bookTitle"`
	Order     int    `json:"order" binding:"required"`
}

func NewSeriesResponse(series Series) SeriesResponse {
	var bookTitles []BooksOrderResponse
	for _, b := range series.Books {
		bookTitles = append(bookTitles, BooksOrderResponse{
			BookTitle: b.Book.Title,
			Order:     b.Order,
		})
	}

	return SeriesResponse{
		ID:    series.ID,
		Title: series.Title,
		Books: bookTitles,
	}
}
