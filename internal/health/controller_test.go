package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestController_Health(t *testing.T) {
	tests := []struct {
		name     string
		expected func(*httptest.ResponseRecorder, error)
	}{
		{
			name: "when successful",
			expected: func(r *httptest.ResponseRecorder, err error) {
				assert.Nil(t, err)
				assert.Equal(t, http.StatusOK, r.Code)
				assert.Equal(t, `{"api":true}`, r.Body.String())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewController()

			r := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(r)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

			c.Health(ctx)

			ctx.Writer.WriteHeaderNow()

			tt.expected(r, ctx.Errors.Last())
		})
	}
}
