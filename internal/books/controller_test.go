package books

import (
	"context"
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
		name         string
		reqBody      string
		setup        func(*ManagerMock)
		expectedCode int
		respBody     string
	}{
		{
			name:         "when request body is an invalid json",
			reqBody:      `}`,
			setup:        func(m *ManagerMock) {},
			expectedCode: http.StatusBadRequest,
			respBody:     `{"error":"invalid request body"}`,
		},
		{
			name:         "when request body is invalid",
			reqBody:      `{}`,
			setup:        func(m *ManagerMock) {},
			expectedCode: http.StatusBadRequest,
			respBody:     `{"error":"Key: 'Book.Title' Error:Field validation for 'Title' failed on the 'required' tag\nKey: 'Book.Year' Error:Field validation for 'Year' failed on the 'required' tag"}`,
		},
		{
			name:    "when create book service fails",
			reqBody: `{"title": "The Black Echo", "year": 1992, "blurb": "a random blurb"}`,
			setup: func(m *ManagerMock) {
				reqBook := Book{Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"}
				m.On("Create", mock.Anything, reqBook).Return(Book{}, assert.AnError).Once()
			},
			expectedCode: http.StatusInternalServerError,
			respBody:     `{"error":"assert.AnError general error for testing"}`,
		},
		{
			name:    "when create book service is successful",
			reqBody: `{"title": "The Black Echo", "year": 1992, "blurb": "a random blurb"}`,
			setup: func(m *ManagerMock) {
				reqBook := Book{Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"}
				respBook := Book{Id: "a-string", Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"}
				m.On("Create", mock.Anything, reqBook).Return(respBook, nil).Once()
			},
			expectedCode: http.StatusCreated,
			respBody:     `{"id":"a-string","title":"The Black Echo","year":1992,"blurb":"a random blurb"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(ManagerMock)
			c := NewController(m, validator.New())

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(tt.reqBody))

			tt.setup(m)

			c.Create(ctx)

			ctx.Writer.WriteHeaderNow()

			assert.Equal(t, tt.expectedCode, recorder.Code)
			assert.Equal(t, tt.respBody, recorder.Body.String())
			m.AssertExpectations(t)
		})
	}
}

func TestController_GetById(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*ManagerMock, *gin.Context)
		expectedCode int
		respBody     string
	}{
		{
			name:         "when missing book id req param",
			setup:        func(_ *ManagerMock, _ *gin.Context) {},
			expectedCode: http.StatusBadRequest,
			respBody:     `{"error":"bookID is required"}`,
		},
		{
			name: "when get book service fails",
			setup: func(m *ManagerMock, ctx *gin.Context) {
				ctx.Params = gin.Params{{Key: "bookID", Value: "a-book-id"}}
				m.On("GetById", mock.Anything, "a-book-id").Return(Book{}, assert.AnError).Once()
			},
			expectedCode: http.StatusInternalServerError,
			respBody:     `{"error":"assert.AnError general error for testing"}`,
		},
		{
			name: "when get book service is successful",
			setup: func(m *ManagerMock, ctx *gin.Context) {
				ctx.Params = gin.Params{{Key: "bookID", Value: "a-book-id"}}
				respBook := Book{Id: "a-string", Title: "The Black Echo", Year: 1992, Blurb: "a random blurb"}
				m.On("GetById", mock.Anything, "a-book-id").Return(respBook, nil).Once()
			},
			expectedCode: http.StatusOK,
			respBody:     `{"id":"a-string","title":"The Black Echo","year":1992,"blurb":"a random blurb"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(ManagerMock)
			c := NewController(m, validator.New())

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/books/book-id", nil)

			tt.setup(m, ctx)

			c.GetById(ctx)

			ctx.Writer.WriteHeaderNow()

			assert.Equal(t, tt.expectedCode, recorder.Code)
			assert.Equal(t, tt.respBody, recorder.Body.String())
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
