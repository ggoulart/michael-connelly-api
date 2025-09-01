package series

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ggoulart/michael-connelly-api/internal/books"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestController_Create(t *testing.T) {
	tests := []struct {
		name     string
		reqBody  string
		setup    func(*ManagerMock)
		expected func(*httptest.ResponseRecorder, error)
	}{
		{
			name:    "when request body is an invalid json",
			reqBody: `}`,
			setup:   func(m *ManagerMock) {},
			expected: func(_ *httptest.ResponseRecorder, err error) {
				var syntaxErr *json.SyntaxError
				assert.True(t, errors.As(err, &syntaxErr))
			},
		},
		{
			name:    "when create series service fails",
			reqBody: `{"title":"The Harry Bosch", "books":[{"title":"The Black Echo", "order": 1}]}`,
			setup: func(m *ManagerMock) {
				reqSeries := Series{Title: "The Harry Bosch"}
				reqOrder := []BooksOrder{{Order: 1, Book: books.Book{Title: "The Black Echo"}}}
				m.On("Create", mock.Anything, reqSeries, reqOrder).Return(Series{}, assert.AnError).Once()
			},
			expected: func(_ *httptest.ResponseRecorder, err error) {
				assert.True(t, errors.Is(err, assert.AnError))
			},
		},
		{
			name:    "when create series is successful",
			reqBody: `{"title":"The Harry Bosch", "books":[{"title":"The Black Echo", "order": 1}]}`,
			setup: func(m *ManagerMock) {
				reqSeries := Series{Title: "The Harry Bosch"}
				reqOrder := []BooksOrder{{Order: 1, Book: books.Book{Title: "The Black Echo"}}}
				outputBook := books.Book{ID: "the-black-echo-book-id", Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"}
				outputSeries := Series{ID: "the-harry-bosch-series-id", Title: "The Harry Bosch", Books: []BooksOrder{{Order: 1, Book: outputBook}}}
				m.On("Create", mock.Anything, reqSeries, reqOrder).Return(outputSeries, nil).Once()
			},
			expected: func(r *httptest.ResponseRecorder, err error) {
				assert.Nil(t, err)
				assert.Equal(t, http.StatusCreated, r.Code)
				assert.Equal(t, `{"id":"the-harry-bosch-series-id","title":"The Harry Bosch","books":[{"id":"the-black-echo-book-id","title":"The Black Echo","order":1}]}`, r.Body.String())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(ManagerMock)
			c := NewController(m)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/series", strings.NewReader(tt.reqBody))

			tt.setup(m)

			c.Create(ctx)

			ctx.Writer.WriteHeaderNow()

			tt.expected(recorder, ctx.Errors.Last())
			m.AssertExpectations(t)
		})
	}
}

func TestController_GetAll(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*ManagerMock)
		expected func(*httptest.ResponseRecorder, error)
	}{
		{
			name: "when failed to get all series",
			setup: func(m *ManagerMock) {
				m.On("GetAll", mock.Anything).Return([]Series{}, assert.AnError).Once()
			},
			expected: func(_ *httptest.ResponseRecorder, err error) {
				assert.True(t, errors.Is(err, assert.AnError))
			},
		},
		{
			name: "when successful to get all series",
			setup: func(m *ManagerMock) {
				series := []Series{{ID: "the-harry-bosch-series-id", Title: "The Harry Bosch", Books: []BooksOrder{{Order: 1, Book: books.Book{ID: "the-black-echo-id", Title: "The Black Echo", Year: 1992, Blurb: "Blurb"}}}}}
				m.On("GetAll", mock.Anything).Return(series, nil).Once()
			},
			expected: func(r *httptest.ResponseRecorder, err error) {
				assert.Nil(t, err)
				assert.Equal(t, http.StatusOK, r.Code)
				assert.Equal(t, `[{"id":"the-harry-bosch-series-id","title":"The Harry Bosch","books":[{"id":"the-black-echo-id","title":"The Black Echo","order":1}]}]`, r.Body.String())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(ManagerMock)
			c := NewController(m)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/series", nil)

			tt.setup(m)

			c.GetAll(ctx)

			ctx.Writer.WriteHeaderNow()

			tt.expected(recorder, ctx.Errors.Last())
			m.AssertExpectations(t)
		})
	}
}

type ManagerMock struct {
	Manager
	mock.Mock
}

func (m *ManagerMock) Create(ctx context.Context, series Series, booksOrderList []BooksOrder) (Series, error) {
	args := m.Called(ctx, series, booksOrderList)
	return args.Get(0).(Series), args.Error(1)
}

func (m *ManagerMock) GetAll(ctx context.Context) ([]Series, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Series), args.Error(1)
}
