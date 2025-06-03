package middleware

import (
	"encoding/json"
	"errors"

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

		switch {
		case errors.As(err, &validationErrs) || errors.As(err, &jsonSyntaxError) || errors.As(err, &jsonUnmarshalTypeError):
			ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		default:
			ctx.AbortWithStatusJSON(500, gin.H{"error": "unexpected error"})
		}
	}
}
