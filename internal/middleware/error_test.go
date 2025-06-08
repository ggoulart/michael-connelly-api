package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ggoulart/michael-connelly-api/internal/dynamo"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*gin.Context)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "when no error",
			setup:          func(ctx *gin.Context) {},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "when error is dynamo.ErrNotFound",
			setup:          func(ctx *gin.Context) { ctx.Error(dynamo.ErrNotFound) },
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"not found"}`,
		},
		{
			name:           "when error is validator.ValidationErrors",
			setup:          func(ctx *gin.Context) { ctx.Error(validator.ValidationErrors{}) },
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":""}`,
		},
		{
			name:           "when error is json.SyntaxError",
			setup:          func(ctx *gin.Context) { ctx.Error(&json.SyntaxError{}) },
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":""}`,
		},
		{
			name: "when error is json.UnmarshalTypeError",
			setup: func(ctx *gin.Context) {
				ctx.Error(&json.UnmarshalTypeError{Field: "title", Value: "string", Type: reflect.TypeOf(123)})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"json: cannot unmarshal string into Go struct field .title of type int"}`,
		},
		{
			name:           "when error is not a known error",
			setup:          func(ctx *gin.Context) { ctx.Error(assert.AnError) },
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"unexpected error"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)

			tt.setup(ctx)

			Error()(ctx)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			assert.Equal(t, tt.expectedBody, recorder.Body.String())
		})
	}
}
