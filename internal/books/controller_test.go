package books

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
			name:    "when create book service fails",
			reqBody: `{"title": "The Black Echo", "year": 1992, "blurb": "a random blurb"}`,
			setup: func(m *ManagerMock) {
				reqBook := Book{Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"}
				m.On("Create", mock.Anything, reqBook).Return(Book{}, assert.AnError).Once()
			},
			expected: func(_ *httptest.ResponseRecorder, err error) {
				assert.True(t, errors.Is(err, assert.AnError))
			},
		},
		{
			name:    "when create book service is successful",
			reqBody: `{"title": "The Black Echo", "year": 1992, "blurb": "a random blurb", "adaptations": [{"description": "Bosch S03","imdb": "https://www.imdb.com/title/tt3502248/episodes/?season=3"}]}`,
			setup: func(m *ManagerMock) {
				reqBook := Book{Title: "The Black Echo", Year: 1992, Blurb: "a random blurb", Adaptations: []Adaptation{{Description: "Bosch S03", IMDB: "https://www.imdb.com/title/tt3502248/episodes/?season=3"}}}
				respBook := Book{ID: "a-string", Title: "The Black Echo", Year: 1992, Blurb: "a random blurb", Adaptations: []Adaptation{{Description: "Bosch S03", IMDB: "https://www.imdb.com/title/tt3502248/episodes/?season=3"}}}
				m.On("Create", mock.Anything, reqBook).Return(respBook, nil).Once()
			},
			expected: func(r *httptest.ResponseRecorder, err error) {
				assert.Nil(t, err)
				assert.Equal(t, http.StatusCreated, r.Code)
				assert.Equal(t, `{"id":"a-string","title":"The Black Echo","year":1992,"blurb":"a random blurb","adaptations":[{"description":"Bosch S03","imdb":"https://www.imdb.com/title/tt3502248/episodes/?season=3"}]}`, r.Body.String())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(ManagerMock)
			c := NewController(m)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(tt.reqBody))

			tt.setup(m)

			c.Create(ctx)

			ctx.Writer.WriteHeaderNow()

			tt.expected(recorder, ctx.Errors.Last())
			m.AssertExpectations(t)
		})
	}
}

func TestController_GetById(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*ManagerMock, *gin.Context)
		expected func(*httptest.ResponseRecorder, error)
	}{
		{
			name:  "when missing book id req param",
			setup: func(_ *ManagerMock, _ *gin.Context) {},
			expected: func(_ *httptest.ResponseRecorder, err error) {
				var validationErrs validator.ValidationErrors
				assert.True(t, errors.As(err, &validationErrs))
			},
		},
		{
			name: "when get book service fails",
			setup: func(m *ManagerMock, ctx *gin.Context) {
				ctx.Params = gin.Params{{Key: "bookID", Value: "a-book-id"}}
				m.On("GetById", mock.Anything, "a-book-id").Return(Book{}, assert.AnError).Once()
			},
			expected: func(_ *httptest.ResponseRecorder, err error) {
				assert.True(t, errors.Is(err, assert.AnError))
			},
		},
		{
			name: "when get book service is successful",
			setup: func(m *ManagerMock, ctx *gin.Context) {
				ctx.Params = gin.Params{{Key: "bookID", Value: "a-book-id"}}
				respBook := Book{ID: "a-string", Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"}
				m.On("GetById", mock.Anything, "a-book-id").Return(respBook, nil).Once()
			},
			expected: func(r *httptest.ResponseRecorder, err error) {
				assert.Nil(t, err)
				assert.Equal(t, http.StatusOK, r.Code)
				assert.Equal(t, `{"id":"a-string","title":"The Black Echo","year":1992,"blurb":"a random blurb"}`, r.Body.String())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(ManagerMock)
			c := NewController(m)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/books/book-id", nil)

			tt.setup(m, ctx)

			c.GetById(ctx)

			ctx.Writer.WriteHeaderNow()

			tt.expected(recorder, ctx.Errors.Last())
			m.AssertExpectations(t)
		})
	}
}

type ManagerMock struct {
	mock.Mock
}

func (m *ManagerMock) Create(ctx context.Context, book Book) (Book, error) {
	args := m.Called(ctx, book)
	return args.Get(0).(Book), args.Error(1)
}

func (m *ManagerMock) GetById(ctx context.Context, bookID string) (Book, error) {
	args := m.Called(ctx, bookID)
	return args.Get(0).(Book), args.Error(1)
}
