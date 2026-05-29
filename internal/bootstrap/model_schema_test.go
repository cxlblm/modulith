package bootstrap

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestModelsUseAutoIncrementIDAndBusinessUUID(t *testing.T) {
	businessModels := map[string]bool{
		"UserModel":          true,
		"AddressModel":       true,
		"ProductModel":       true,
		"OrderModel":         true,
		"PaymentModel":       true,
		"ShipmentModel":      true,
		"QuestionModel":      true,
		"ContestModel":       true,
		"ParticipationModel": true,
	}

	for _, model := range allModels() {
		modelType := reflect.TypeOf(model)
		if modelType.Kind() != reflect.Pointer {
			t.Fatalf("model %T is not a pointer", model)
		}
		modelType = modelType.Elem()
		t.Run(modelType.Name(), func(t *testing.T) {
			idField, ok := modelType.FieldByName("ID")
			if !ok {
				t.Fatalf("%s has no ID field", modelType.Name())
			}
			if idField.Type.Kind() != reflect.Uint64 {
				t.Fatalf("%s.ID type = %s, want uint64", modelType.Name(), idField.Type)
			}
			assertGORMTagContains(t, idField, "primaryKey")
			assertGORMTagContains(t, idField, "autoIncrement")

			if !businessModels[modelType.Name()] {
				return
			}
			uuidField, ok := modelType.FieldByName("UUID")
			if !ok {
				t.Fatalf("%s has no UUID field", modelType.Name())
			}
			if uuidField.Type.Kind() != reflect.String {
				t.Fatalf("%s.UUID type = %s, want string", modelType.Name(), uuidField.Type)
			}
			assertGORMTagContains(t, uuidField, "not null")
			assertGORMTagContains(t, uuidField, "uniqueIndex")
		})
	}
}

func TestPersistenceReferencesUseUUIDNames(t *testing.T) {
	expectedFields := map[string][]string{
		"StockReservationModel":    {"ProductUUID", "OrderUUID"},
		"OrderItemModel":           {"OrderUUID", "ProductUUID"},
		"PaymentModel":             {"OrderUUID"},
		"ShipmentModel":            {"OrderUUID"},
		"QuestionOptionModel":      {"QuestionUUID"},
		"ContestQuestionModel":     {"ContestUUID", "QuestionUUID"},
		"ParticipationModel":       {"ContestUUID", "UserUUID"},
		"ParticipationAnswerModel": {"ParticipationUUID", "QuestionUUID"},
		"RevivalCardModel":         {"UserUUID"},
		"RewardClaimModel":         {"ContestUUID", "UserUUID"},
	}

	for _, model := range allModels() {
		modelType := reflect.TypeOf(model)
		if modelType.Kind() != reflect.Pointer {
			t.Fatalf("model %T is not a pointer", model)
		}
		modelType = modelType.Elem()
		wantFields, ok := expectedFields[modelType.Name()]
		if !ok {
			continue
		}
		t.Run(modelType.Name(), func(t *testing.T) {
			if _, ok := modelType.FieldByName("ProductID"); ok {
				t.Fatalf("%s has ProductID field, want ProductUUID", modelType.Name())
			}
			if _, ok := modelType.FieldByName("OrderID"); ok {
				t.Fatalf("%s has OrderID field, want OrderUUID", modelType.Name())
			}
			for _, fieldName := range wantFields {
				field, ok := modelType.FieldByName(fieldName)
				if !ok {
					t.Fatalf("%s has no %s field", modelType.Name(), fieldName)
				}
				if field.Type.Kind() != reflect.String {
					t.Fatalf("%s.%s type = %s, want string", modelType.Name(), fieldName, field.Type)
				}
				assertGORMTagContains(t, field, "not null")
			}
		})
	}
}

func TestPersistenceModelsHaveAuditTimestamps(t *testing.T) {
	for _, model := range allModels() {
		modelType := reflect.TypeOf(model)
		if modelType.Kind() != reflect.Pointer {
			t.Fatalf("model %T is not a pointer", model)
		}
		modelType = modelType.Elem()
		t.Run(modelType.Name(), func(t *testing.T) {
			assertTimeField(t, modelType, "CreatedAt")
			assertTimeField(t, modelType, "UpdatedAt")
		})
	}
}

func assertGORMTagContains(t *testing.T, field reflect.StructField, want string) {
	t.Helper()
	tag := field.Tag.Get("gorm")
	for _, part := range strings.Split(tag, ";") {
		if part == want || strings.HasPrefix(part, want+":") {
			return
		}
	}
	t.Fatalf("%s gorm tag = %q, want %q", field.Name, tag, want)
}

func assertTimeField(t *testing.T, modelType reflect.Type, fieldName string) {
	t.Helper()
	field, ok := modelType.FieldByName(fieldName)
	if !ok {
		t.Fatalf("%s has no %s field", modelType.Name(), fieldName)
	}
	if field.Type != reflect.TypeOf(time.Time{}) {
		t.Fatalf("%s.%s type = %s, want time.Time", modelType.Name(), fieldName, field.Type)
	}
}
