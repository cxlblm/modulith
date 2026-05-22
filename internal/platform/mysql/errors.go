package mysql

import (
	"errors"

	drivermysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var mysqlErr *drivermysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
