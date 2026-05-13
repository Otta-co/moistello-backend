package validator

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func Init() {
	Validate = validator.New()
	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		*v = *Validate
	}
}

func FormatError(err error) []map[string]string {
	var errors []map[string]string
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			errors = append(errors, map[string]string{
				"field":   e.Field(),
				"message": formatMessage(e),
			})
		}
	}
	return errors
}

func formatMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Must be at least " + e.Param()
	case "max":
		return "Must be at most " + e.Param()
	case "email":
		return "Must be a valid email"
	case "oneof":
		return "Must be one of: " + e.Param()
	case "gt":
		return "Must be greater than " + e.Param()
	case "gte":
		return "Must be at least " + e.Param()
	case "lte":
		return "Must be at most " + e.Param()
	default:
		return e.Tag()
	}
}
