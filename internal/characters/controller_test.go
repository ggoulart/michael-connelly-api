package characters

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ggoulart/michael-connelly-api/internal/books"
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
			name:    "when create character service fails",
			reqBody: `{"name":"Harry Bosch"}`,
			setup: func(m *ManagerMock) {
				reqCharacter := Character{Name: "Harry Bosch"}
				m.On("Create", mock.Anything, reqCharacter, []string(nil)).Return(Character{}, assert.AnError).Once()
			},
			expected: func(_ *httptest.ResponseRecorder, err error) {
				assert.True(t, errors.Is(err, assert.AnError))
			},
		},
		{
			name:    "when create character is successful",
			reqBody: `{"name":"Harry Bosch", "bookTitles": ["random-book-title"]}`,
			setup: func(m *ManagerMock) {
				reqCharacter := Character{Name: "Harry Bosch"}
				respCharacter := Character{ID: "random-id", Name: "Harry Bosch", Books: []books.Book{{Title: "random-book-title"}}}
				m.On("Create", mock.Anything, reqCharacter, []string{"random-book-title"}).Return(respCharacter, nil).Once()
			},
			expected: func(r *httptest.ResponseRecorder, err error) {
				assert.Nil(t, err)
				assert.Equal(t, http.StatusCreated, r.Code)
				assert.Equal(t, `{"id":"random-id","name":"Harry Bosch","booksTitles":["random-book-title"]}`, r.Body.String())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(ManagerMock)
			c := NewController(m)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/characters", strings.NewReader(tt.reqBody))

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
		setup    func(*gin.Context, *ManagerMock)
		expected func(*httptest.ResponseRecorder, error)
	}{
		{
			name:  "when missing character id",
			setup: func(_ *gin.Context, _ *ManagerMock) {},
			expected: func(_ *httptest.ResponseRecorder, err error) {
				var validationErrs validator.ValidationErrors
				assert.True(t, errors.As(err, &validationErrs))
			},
		},
		{
			name: "when get character service fails",
			setup: func(ctx *gin.Context, m *ManagerMock) {
				ctx.Params = gin.Params{{Key: "character", Value: "c6767b2d-438b-4d4c-8b1a-659130a640ca"}}
				m.On("GetById", mock.Anything, "c6767b2d-438b-4d4c-8b1a-659130a640ca").Return(Character{}, assert.AnError).Once()
			},
			expected: func(_ *httptest.ResponseRecorder, err error) {
				assert.True(t, errors.Is(err, assert.AnError))
			},
		},
		{
			name: "when get character service is successful",
			setup: func(ctx *gin.Context, m *ManagerMock) {
				ctx.Params = gin.Params{{Key: "character", Value: "c6767b2d-438b-4d4c-8b1a-659130a640ca"}}
				respCharacter := Character{ID: "c6767b2d-438b-4d4c-8b1a-659130a640ca", Name: "Harry Bosch"}
				m.On("GetById", mock.Anything, "c6767b2d-438b-4d4c-8b1a-659130a640ca").Return(respCharacter, nil).Once()
			},
			expected: func(r *httptest.ResponseRecorder, err error) {
				assert.Nil(t, err)
				assert.Equal(t, http.StatusOK, r.Code)
				assert.Equal(t, `{"id":"c6767b2d-438b-4d4c-8b1a-659130a640ca","name":"Harry Bosch"}`, r.Body.String())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(ManagerMock)
			c := NewController(m)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/characters/c6767b2d-438b-4d4c-8b1a-659130a640ca", nil)

			tt.setup(ctx, m)

			c.GetBy(ctx)

			ctx.Writer.WriteHeaderNow()

			tt.expected(recorder, ctx.Errors.Last())
			m.AssertExpectations(t)
		})
	}
}

func TestController_GetByName(t *testing.T) {
	tests := []struct {
		name           string
		nameQueryParam string
		setup          func(*gin.Context, *ManagerMock)
		expected       func(*httptest.ResponseRecorder, error)
	}{
		{
			name:  "when missing character name",
			setup: func(_ *gin.Context, _ *ManagerMock) {},
			expected: func(_ *httptest.ResponseRecorder, err error) {
				var validationErrs validator.ValidationErrors
				assert.True(t, errors.As(err, &validationErrs))
			},
		},
		{
			name:           "when get character service fails",
			nameQueryParam: "Harry Bosch",
			setup: func(ctx *gin.Context, m *ManagerMock) {
				ctx.Params = gin.Params{{Key: "character", Value: "Harry Bosch"}}
				m.On("GetByName", mock.Anything, "Harry Bosch").Return(Character{}, assert.AnError).Once()
			},
			expected: func(_ *httptest.ResponseRecorder, err error) {
				assert.True(t, errors.Is(err, assert.AnError))
			},
		},
		{
			name:           "when get character service is successful",
			nameQueryParam: "Harry Bosch",
			setup: func(ctx *gin.Context, m *ManagerMock) {
				ctx.Params = gin.Params{{Key: "character", Value: "Harry Bosch"}}
				character := Character{ID: "random-id", Name: "Harry Bosch"}
				m.On("GetByName", mock.Anything, "Harry Bosch").Return(character, nil).Once()
			},
			expected: func(r *httptest.ResponseRecorder, err error) {
				assert.Nil(t, err)
				assert.Equal(t, http.StatusOK, r.Code)
				assert.Equal(t, `{"id":"random-id","name":"Harry Bosch"}`, r.Body.String())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(ManagerMock)
			c := NewController(m)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/characters/"+url.QueryEscape("Harry Bosch"), nil)

			tt.setup(ctx, m)

			c.GetBy(ctx)

			ctx.Writer.WriteHeaderNow()

			tt.expected(recorder, ctx.Errors.Last())
			m.AssertExpectations(t)
		})
	}
}

type ManagerMock struct {
	mock.Mock
}

func (m *ManagerMock) GetByName(ctx context.Context, characterName string) (Character, error) {
	args := m.Called(ctx, characterName)
	return args.Get(0).(Character), args.Error(1)
}

func (m *ManagerMock) Create(ctx context.Context, character Character, bookTitle []string) (Character, error) {
	args := m.Called(ctx, character, bookTitle)
	return args.Get(0).(Character), args.Error(1)
}

func (m *ManagerMock) GetById(ctx context.Context, characterID string) (Character, error) {
	args := m.Called(ctx, characterID)
	return args.Get(0).(Character), args.Error(1)
}
