package utils

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ValidateBody(context *gin.Context, payload interface{}) bool {

	if err := context.ShouldBind(&payload); err != nil {
		var validationErrors validator.ValidationErrors

		if errors.As(err, &validationErrors) {
			context.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Validation error", "errors": errorsFormat(validationErrors)})
			return false
		}

		// We now know that this error is not a validation error
		// probably a malformed JSON

		context.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "bad request"})
		return false
	}

	return true

}

func errorsFormat(validationErrors validator.ValidationErrors) map[string]string {
	errs := make(map[string]string)

	for _, f := range validationErrors {
		err := f.ActualTag()
		if f.Param() != "" {
			err = fmt.Sprintf("%s=%s", err, f.Param())
		}

		err = strings.Replace(err, "required", "required field", -1)

		errs[camelToSnake(f.Field())] = err
	}

	return errs
}

func camelToSnake(input string) string {
	re := regexp.MustCompile("([a-z])([A-Z])")
	snake := re.ReplaceAllString(input, "${1}_${2}")
	return strings.ToLower(snake)
}
