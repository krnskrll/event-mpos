package common

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

// BindingMessage renders a request-binding error into a "field: reason; field: reason" string,
// approximating the Kotlin MethodArgumentNotValidException message. Falls back to the raw error.
func BindingMessage(err error) string {
	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		if err != nil {
			return err.Error()
		}
		return "Validation failed"
	}
	parts := make([]string, 0, len(ve))
	for _, fe := range ve {
		field := lowerFirst(fe.Field())
		var reason string
		switch fe.Tag() {
		case "required":
			reason = "must not be blank"
		case "email":
			reason = "must be a well-formed email address"
		case "min":
			reason = "must be at least " + fe.Param() + " characters"
		default:
			reason = "is invalid"
		}
		parts = append(parts, field+": "+reason)
	}
	return strings.Join(parts, "; ")
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
