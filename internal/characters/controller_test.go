package characters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ggoulart/michael-connelly-api/internal/books"
	"github.com/gin-gonic/gin"
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
			name:    "when create character service fails",
			reqBody: `{"name":"Harry Bosch"}`,
			setup: func(m *ManagerMock) {
				reqCharacter := Character{Name: "Harry Bosch"}
				m.On("Create", mock.Anything, reqCharacter).Return(Character{}, assert.AnError).Once()
			},
			expectedCode: http.StatusInternalServerError,
			respBody:     `{"error":"assert.AnError general error for testing"}`,
		},
		{
			name:    "when create character is successful",
			reqBody: `{"name":"Harry Bosch"}`,
			setup: func(m *ManagerMock) {
				reqCharacter := Character{Name: "Harry Bosch"}
				respCharacter := Character{ID: "random-id", Name: "Harry Bosch", Books: []books.Book{{Title: "random-book-title"}}}
				m.On("Create", mock.Anything, reqCharacter).Return(respCharacter, nil).Once()
			},
			expectedCode: http.StatusCreated,
			respBody:     `{"id":"random-id","name":"Harry Bosch","booksTitles":["random-book-title"]}`,
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

			assert.Equal(t, tt.expectedCode, recorder.Code)
			assert.Equal(t, tt.respBody, recorder.Body.String())
			m.AssertExpectations(t)
		})
	}
}

func TestController_GetById(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*gin.Context, *ManagerMock)
		expectedCode int
		respBody     string
	}{
		{
			name:         "when missing character id req param",
			setup:        func(_ *gin.Context, _ *ManagerMock) {},
			expectedCode: http.StatusBadRequest,
			respBody:     `{"error":"characterID is required"}`,
		},
		{
			name: "when get character service fails",
			setup: func(ctx *gin.Context, m *ManagerMock) {
				ctx.Params = gin.Params{{Key: "characterID", Value: "a-character-id"}}
				m.On("GetById", mock.Anything, "a-character-id").Return(Character{}, assert.AnError).Once()
			},
			expectedCode: http.StatusInternalServerError,
			respBody:     `{"error":"assert.AnError general error for testing"}`,
		},
		{
			name: "when get character service is successful",
			setup: func(ctx *gin.Context, m *ManagerMock) {
				ctx.Params = gin.Params{{Key: "characterID", Value: "a-character-id"}}
				respCharacter := Character{ID: "a-character-id", Name: "Harry Bosch"}
				m.On("GetById", mock.Anything, "a-character-id").Return(respCharacter, nil).Once()
			},
			expectedCode: http.StatusOK,
			respBody:     `{"id":"a-character-id","name":"Harry Bosch"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(ManagerMock)
			c := NewController(m)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/characters/character-id", nil)

			tt.setup(ctx, m)

			c.GetById(ctx)

			ctx.Writer.WriteHeaderNow()

			assert.Equal(t, tt.expectedCode, recorder.Code)
			assert.Equal(t, tt.respBody, recorder.Body.String())
			m.AssertExpectations(t)
		})
	}
}

func TestController_GetByName(t *testing.T) {
	tests := []struct {
		name           string
		nameQueryParam string
		setup          func(*ManagerMock)
		expectedCode   int
		respBody       string
	}{
		{
			name:         "when missing character name req param",
			setup:        func(_ *ManagerMock) {},
			expectedCode: http.StatusBadRequest,
			respBody:     `{"error":"query param 'name' is required"}`,
		},
		{
			name:           "when get character service fails",
			nameQueryParam: "Harry Bosch",
			setup: func(m *ManagerMock) {
				m.On("GetByName", mock.Anything, "Harry Bosch").Return(Character{}, assert.AnError).Once()
			},
			expectedCode: http.StatusInternalServerError,
			respBody:     `{"error":"assert.AnError general error for testing"}`,
		},
		{
			name:           "when get character service is successful",
			nameQueryParam: "Harry Bosch",
			setup: func(m *ManagerMock) {
				character := Character{ID: "random-id", Name: "Harry Bosch"}
				m.On("GetByName", mock.Anything, "Harry Bosch").Return(character, nil).Once()
			},
			expectedCode: http.StatusOK,
			respBody:     `{"id":"random-id","name":"Harry Bosch"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(ManagerMock)
			c := NewController(m)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/characters/?name="+url.QueryEscape(tt.nameQueryParam), nil)

			tt.setup(m)

			c.GetByName(ctx)

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

func (m *ManagerMock) GetByName(ctx context.Context, characterName string) (Character, error) {
	args := m.Called(ctx, characterName)
	return args.Get(0).(Character), args.Error(1)
}

func (m *ManagerMock) Create(ctx context.Context, character Character) (Character, error) {
	args := m.Called(ctx, character)
	return args.Get(0).(Character), args.Error(1)
}

func (m *ManagerMock) GetById(ctx context.Context, characterID string) (Character, error) {
	args := m.Called(ctx, characterID)
	return args.Get(0).(Character), args.Error(1)
}
