package app_validation

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

type AppError struct {
	code    string
	key     string
	comment string
}

func (err AppError) Error() string {
	return strings.Trim(err.code+" : "+err.key+" - "+err.comment, " -:")
}

type AppErrorCollection struct {
	Errors []AppError
}

func (err AppErrorCollection) Error() string {
	if err.HasError() {
		suffix := ""
		if len(err.Errors) > 1 {
			suffix = "..."
		}
		return err.Errors[0].Error() + suffix
	}
	return ""
}

func (err AppErrorCollection) HasError() bool {
	return err.Errors != nil && len(err.Errors) > 0
}

func (err *AppErrorCollection) Append(e AppError) {
	err.Errors = append(err.Errors, e)
}
func (err *AppErrorCollection) Merge(e AppErrorCollection) {
	err.Errors = append(err.Errors, e.Errors...)
}

func NewAppError(code, key, comment string) *AppError {
	return &AppError{code, key, comment}
}

func (err AppError) Code() string {
	return err.code
}

func (err AppError) Key() string {
	return err.key
}

func (err AppError) Comment() string {
	return err.comment
}

func (err AppError) SetKey(key string) *AppError {
	err.key = key
	return &err
}

func (err AppError) SetComment(comment string) *AppError {
	err.comment = comment
	return &err
}

// TODO: move this to new Package called app_errors
var (
	ErrInvalid   = NewAppError("invalid", "", "please fill with valid options")
	ErrRequired  = NewAppError("required", "", "please fill required field")
	ErrNotSet    = NewAppError("not_set", "", "")
	ErrNotExists = NewAppError("not_exists", "", "")
	ErrUnique    = NewAppError("unique", "", "")
	ErrExpired   = NewAppError("expired", "", "")
	ErrNotfound  = NewAppError("not_found", "", "")
	ErrNotPaid   = NewAppError("not_paid", "", "")

	ErrUnauthorized = NewAppError("unauthorized", "user", "")
	ErrForbidden    = NewAppError("forbidden", "user", "")
)

func ValidateStruct(r interface{}) error {
	v := validator.New()
	errs := AppErrorCollection{}
	if err := v.Struct(r); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errs.Errors = append(errs.Errors, *NewAppError(err.Tag(), pascalCaseToSnakeCase(err.Field()), err.Error()))
		}
		return errs
	}
	return nil
}

func pascalCaseToSnakeCase(str string) string {
	snake := []rune{}
	isUpper := false
	for _, char := range str {
		if char >= 'A' && char <= 'Z' {
			if !isUpper {
				snake = append(snake, '_')
			}
			isUpper = true
			snake = append(snake, rune(char+32)) // Convert to lowercase
		} else {
			snake = append(snake, char)
			isUpper = false
		}
	}
	return strings.Trim(string(snake), " _")
}
