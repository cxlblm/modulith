package mysql

import (
	"fmt"
	"testing"

	drivermysql "github.com/go-sql-driver/mysql"
)

func TestIsDuplicateKeyDetectsWrappedMySQLError1062(t *testing.T) {
	err := fmt.Errorf("create model: %w", &drivermysql.MySQLError{Number: 1062, Message: "duplicate entry"})
	if !IsDuplicateKey(err) {
		t.Fatal("IsDuplicateKey() = false, want true")
	}
}

func TestIsDuplicateKeyRejectsOtherMySQLError(t *testing.T) {
	err := &drivermysql.MySQLError{Number: 1452, Message: "foreign key constraint fails"}
	if IsDuplicateKey(err) {
		t.Fatal("IsDuplicateKey() = true, want false")
	}
}
