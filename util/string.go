package util

import (
	"github.com/google/uuid"
)

func MustUUID() string {
	_uuid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return _uuid.String()
}
