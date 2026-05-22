package util

func IIf[T any](condition bool, trueVal T, falseVal T) T {
	if condition {
		return trueVal
	} else {
		return falseVal
	}
}
