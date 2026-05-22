package validation

import (
	"errors"
	"testing"
)

func TestValidator_Validate_ReturnsJSONFieldNames(t *testing.T) {
	type createUserRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	err := New().Validate(createUserRequest{})
	if err == nil {
		t.Fatal("Validate() error = nil, want validation error")
	}

	validationErr, ok := errors.AsType[Error](err)
	if !ok {
		t.Fatalf("Validate() error type = %T, want validation.Error", err)
	}

	fields := validationErr.Fields
	if len(fields) != 2 {
		t.Fatalf("len(fields) = %d, want %d", len(fields), 2)
	}
	if fields[0].Field != "name" || fields[0].Rule != "required" {
		t.Fatalf("fields[0] = %#v, want field name required", fields[0])
	}
	if fields[1].Field != "email" || fields[1].Rule != "required" {
		t.Fatalf("fields[1] = %#v, want field email required", fields[1])
	}
}

func TestValidator_Validate_AllowsValidStruct(t *testing.T) {
	type createUserRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	err := New().Validate(createUserRequest{
		Name:  "Ada",
		Email: "ada@example.com",
	})
	if err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}
