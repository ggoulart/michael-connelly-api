package series

import (
	"context"
	"net/http"

	"github.com/ggoulart/michael-connelly-api/internal/books"
	"github.com/gin-gonic/gin"
)

type Manager interface {
	Create(ctx context.Context, series Series, booksOrderList []BooksOrder) (Series, error)
	GetAll(ctx context.Context) ([]Series, error)
}

type Controller struct {
	manager Manager
}

func NewController(manager Manager) *Controller {
	return &Controller{manager: manager}
}

func (c *Controller) Create(ctx *gin.Context) {
	var seriesDTO SeriesDTO
	if err := ctx.BindJSON(&seriesDTO); err != nil {
		ctx.Error(err)
		return
	}

	createdSeries, err := c.manager.Create(ctx, seriesDTO.ToSeries(), seriesDTO.ToBooksOrderList())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, NewSeriesDTO(createdSeries))
}

func (c *Controller) GetAll(ctx *gin.Context) {
	series, err := c.manager.GetAll(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var seriesDTO []SeriesDTO
	for _, s := range series {
		seriesDTO = append(seriesDTO, NewSeriesDTO(s))
	}

	ctx.JSON(http.StatusOK, seriesDTO)

}

type SeriesDTO struct {
	ID    string          `json:"id"`
	Title string          `json:"title" binding:"required"`
	Books []BooksOrderDTO `json:"books"`
}

type BooksOrderDTO struct {
	ID        string `json:"id,omitempty"`
	BookTitle string `json:"title" binding:"required"`
	Order     int    `json:"order" binding:"required"`
}

func NewSeriesDTO(series Series) SeriesDTO {
	var bookTitles []BooksOrderDTO
	for _, b := range series.Books {
		bookTitles = append(bookTitles, BooksOrderDTO{
			ID:        b.Book.ID,
			BookTitle: b.Book.Title,
			Order:     b.Order,
		})
	}

	return SeriesDTO{
		ID:    series.ID,
		Title: series.Title,
		Books: bookTitles,
	}
}

func (r *SeriesDTO) ToSeries() Series {
	return Series{
		Title: r.Title,
	}
}

func (r *SeriesDTO) ToBooksOrderList() []BooksOrder {
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
