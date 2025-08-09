package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestController_Health(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*ManagerMock)
		expected func(*httptest.ResponseRecorder, error)
	}{
		{
			name: "when successful",
			setup: func(m *ManagerMock) {
				m.On("Health", mock.Anything).Return(map[string]bool{"api": true})
			},
			expected: func(r *httptest.ResponseRecorder, err error) {
				assert.Nil(t, err)
				assert.Equal(t, http.StatusOK, r.Code)
				assert.Equal(t, `{"api":true}`, r.Body.String())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(ManagerMock)
			tt.setup(m)
			c := NewController(m)

			r := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(r)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

			c.Health(ctx)

			ctx.Writer.WriteHeaderNow()

			tt.expected(r, ctx.Errors.Last())
		})
	}
}

type ManagerMock struct {
	mock.Mock
}

func (m *ManagerMock) Health(ctx context.Context) map[string]bool {
	args := m.Called(ctx)
	return args.Get(0).(map[string]bool)
}
