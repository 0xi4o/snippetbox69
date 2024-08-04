package validator

import (
	"slices"
	"strings"
	"unicode/utf8"
)

type Validator struct {
	FieldErrors map[string]string
}

func (validator *Validator) Valid() bool {
	return len(validator.FieldErrors) == 0
}

func (validator *Validator) AddFieldError(key, message string) {
	if validator.FieldErrors == nil {
		validator.FieldErrors = make(map[string]string)
	}

	if _, exists := validator.FieldErrors[key]; !exists {
		validator.FieldErrors[key] = message
	}
}

func (validator *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		validator.AddFieldError(key, message)
	}
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MaxChars(value string, max int) bool {
	return utf8.RuneCountInString(value) <= max
}

func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}
