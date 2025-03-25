package util

import (
	"crypto/sha256"
	"fmt"
	"unicode/utf8"

	"github.com/google/uuid"
)

func MustUUID() string {
	_uuid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return _uuid.String()
}

func TruncateString(s string, length int) string {
	if utf8.RuneCountInString(s) > length {
		return string([]rune(s)[:length]) + "..."
	}
	return s
}

func Hash(x string) string {
	hash := sha256.New()
	hash.Write([]byte(x))

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func GenerateUniqKeys(n int) []string {
	keys := make([]string, n)
	for i := range n {
		keys[i] = MustUUID()
	}
	return keys
}
