package util

import (
	"encoding/json"
	"fmt"
	"os"
)

func MustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return b
}

func WriteJson(filename string, data any) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to write JSON data: %w", err)
	}

	return nil
}

func LoadJson[T any](filename string) (*T, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var data T
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON data: %w", err)
	}

	return &data, nil
}

func MustJsonString(j any) string {
	byts, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}

	return string(byts)
}
