package util

import (
	"time"
)

func IsClosed[T any, C interface{ ~chan T | ~<-chan T }](ch C) bool {
	var ok bool
	select {
	case _, ok = <-ch:
	default:
		ok = true
	}
	return !ok
}

func SendOrTimeout[T any, C interface{ ~chan T }](ch C, v T, d time.Duration) bool {
	t := time.NewTicker(d)
	defer t.Stop()
	select {
	case ch <- v:
		return true
	case <-t.C:
		return false
	}
}
