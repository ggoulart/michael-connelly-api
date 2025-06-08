package middleware

import (
	"encoding/json"
	"errors"

	"github.com/ggoulart/michael-connelly-api/internal/dynamo"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func Error() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) == 0 {
			return
		}

		err := ctx.Errors.Last()

		var validationErrs validator.ValidationErrors
		var jsonSyntaxError *json.SyntaxError
		var jsonUnmarshalTypeError *json.UnmarshalTypeError

		//slog.Error(err.Error())

		switch {
		case errors.Is(err, dynamo.ErrNotFound):
			ctx.AbortWithStatusJSON(404, gin.H{"error": "not found"})
		case errors.As(err, &validationErrs) || errors.As(err, &jsonSyntaxError) || errors.As(err, &jsonUnmarshalTypeError):
			ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		default:
			ctx.AbortWithStatusJSON(500, gin.H{"error": "unexpected error"})
		}
	}
}
