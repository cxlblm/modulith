package bizid

import (
	"fmt"

	"github.com/google/uuid"
)

func New() string {
	id, err := uuid.NewV7()
	if err != nil {
		panic(fmt.Sprintf("generate uuid v7: %v", err))
	}
	return id.String()
}
