package query

import (
	"reflect"
	"testing"
	"time"
)

func TestDTOsExposeAuditTimestamps(t *testing.T) {
	assertAuditTimestampFields(t, reflect.TypeOf(OrderDTO{}))
	assertAuditTimestampFields(t, reflect.TypeOf(OrderItemDTO{}))
}

func assertAuditTimestampFields(t *testing.T, dtoType reflect.Type) {
	t.Helper()
	assertAuditTimestampField(t, dtoType, "CreatedAt", "created_at")
	assertAuditTimestampField(t, dtoType, "UpdatedAt", "updated_at")
}

func assertAuditTimestampField(t *testing.T, dtoType reflect.Type, fieldName string, jsonName string) {
	t.Helper()
	field, ok := dtoType.FieldByName(fieldName)
	if !ok {
		t.Fatalf("%s has no %s field", dtoType.Name(), fieldName)
	}
	if field.Type != reflect.TypeOf(time.Time{}) {
		t.Fatalf("%s.%s type = %s, want time.Time", dtoType.Name(), fieldName, field.Type)
	}
	if got := field.Tag.Get("json"); got != jsonName {
		t.Fatalf("%s.%s json tag = %q, want %q", dtoType.Name(), fieldName, got, jsonName)
	}
}
