package util

func IsClosed[T any, C interface{ ~chan T | ~<-chan T }](ch C) bool {
	var ok bool
	select {
	case _, ok = <-ch:
	default:
		ok = true
	}
	return !ok
}
