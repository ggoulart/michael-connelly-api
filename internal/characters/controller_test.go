package characters

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
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
			respBody:     `{"error": "invalid request body"}`,
		},
		{
			name:         "when request body is invalid",
			reqBody:      `{}`,
			setup:        func(m *ManagerMock) {},
			expectedCode: http.StatusBadRequest,
			respBody:     `{"error": "Key: 'Character.Name' Error:Field validation for 'Name' failed on the 'required' tag"}`,
		},
		{
			name:    "when create character service fails",
			reqBody: `{"name":"Harry Bosch"}`,
			setup: func(m *ManagerMock) {
				reqCharacter := Character{Name: "Harry Bosch"}
				m.On("Create", mock.Anything, reqCharacter).Return(Character{}, assert.AnError).Once()
			},
			expectedCode: http.StatusInternalServerError,
			respBody:     `{"error": "unexpected error"}`,
		},
		{
			name:    "when create character is successful",
			reqBody: `{"name":"Harry Bosch"}`,
			setup: func(m *ManagerMock) {
				reqCharacter := Character{Name: "Harry Bosch"}
				respCharacter := Character{Id: "random-id", Name: "Harry Bosch", Books: []string{"random-book-id"}}
				m.On("Create", mock.Anything, reqCharacter).Return(respCharacter, nil).Once()
			},
			expectedCode: http.StatusCreated,
			respBody:     `{"id":"random-id","name":"Harry Bosch","books":["random-book-id"]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(ManagerMock)
			c := NewController(m, validator.New())

			tt.setup(m)

			req := httptest.NewRequest(http.MethodPost, "/characters", strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()

			c.Create(w, req)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			assert.Equal(t, strings.TrimSpace(tt.respBody), strings.TrimSpace(string(body)))
			m.AssertExpectations(t)
		})
	}
}

func TestController_GetById(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*ManagerMock, *chi.Context)
		expectedCode int
		respBody     string
	}{
		{
			name:         "when missing character id req param",
			setup:        func(_ *ManagerMock, _ *chi.Context) {},
			expectedCode: http.StatusBadRequest,
			respBody:     `{"error": "character id is required"}`,
		},
		{
			name: "when get character service fails",
			setup: func(m *ManagerMock, rc *chi.Context) {
				rc.URLParams.Add("characterID", "a-character-id")
				m.On("GetById", mock.Anything, "a-character-id").Return(Character{}, assert.AnError).Once()
			},
			expectedCode: http.StatusInternalServerError,
			respBody:     `{"error": "unexpected error"}`,
		},
		{
			name: "when get character service is successful",
			setup: func(m *ManagerMock, rc *chi.Context) {
				rc.URLParams.Add("characterID", "a-character-id")
				respCharacter := Character{Id: "a-character-id", Name: "Harry Bosch"}
				m.On("GetById", mock.Anything, "a-character-id").Return(respCharacter, nil).Once()
			},
			expectedCode: http.StatusOK,
			respBody:     `{"id":"a-character-id","name":"Harry Bosch"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(ManagerMock)
			c := NewController(m, validator.New())

			req := httptest.NewRequest(http.MethodGet, "/characters/character-id", nil)
			rc := chi.NewRouteContext()
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))
			w := httptest.NewRecorder()

			tt.setup(m, rc)

			c.GetById(w, req)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			assert.Equal(t, strings.TrimSpace(tt.respBody), strings.TrimSpace(string(body)))
			m.AssertExpectations(t)
		})
	}
}

type ManagerMock struct {
	mock.Mock
}

func (m *ManagerMock) Create(ctx context.Context, character Character) (Character, error) {
	args := m.Called(ctx, character)
	return args.Get(0).(Character), args.Error(1)
}

func (m *ManagerMock) GetById(ctx context.Context, characterID string) (Character, error) {
	args := m.Called(ctx, characterID)
	return args.Get(0).(Character), args.Error(1)
}
