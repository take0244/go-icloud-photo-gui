package aop

import "os"

func IsDebug() bool {
	return os.Getenv("DEVELOPMENT") == "true"
}
