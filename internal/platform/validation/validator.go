package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

type Error struct {
	Fields []FieldError `json:"fields"`
}

func (e Error) Error() string {
	return "validation failed"
}

type FieldError struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
}

func New() *Validator {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(jsonFieldName)
	return &Validator{validate: validate}
}

func (v *Validator) Validate(value any) error {
	if err := v.validate.Struct(value); err != nil {
		validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
		if !ok {
			return fmt.Errorf("validate struct: %w", err)
		}
		return Error{Fields: toFieldErrors(validationErrors)}
	}
	return nil
}

func toFieldErrors(validationErrors validator.ValidationErrors) []FieldError {
	fields := make([]FieldError, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		fields = append(fields, FieldError{
			Field: fieldErr.Field(),
			Rule:  fieldErr.Tag(),
		})
	}
	return fields
}

func jsonFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}

	name, _, _ := strings.Cut(tag, ",")
	if name == "" || name == "-" {
		return field.Name
	}
	return name
}
