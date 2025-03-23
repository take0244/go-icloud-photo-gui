package util

import (
	"encoding/json"
)

func MustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return b
}

func MustUnmarshal[T any](v []byte) T {
	t, err := Unmarshal[T](v)
	if err != nil {
		panic(err)
	}

	return *t
}

func Unmarshal[T any](v []byte) (*T, error) {
	var t T
	err := json.Unmarshal(v, &t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func MustJsonString(j any) string {
	byts, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}

	return string(byts)
}
