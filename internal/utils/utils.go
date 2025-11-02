package utils

import (
	"log/slog"
)

var Logger *slog.Logger

// entfernt ntes Element aus slice
func RemoveElementFromSlice[T any](slice []T, n int) []T {
	if n >= len(slice) {
		Logger.Debug("Index out of Bounds. Tried to remove the not existing Element " + string(rune(n)))
	}
	if n < 0 {
		Logger.Debug("Index out of Bounds. Tried to remove a negative Element" + string(rune(n)))
	}
	tempSlice := slice
	first := tempSlice[:n]
	var last []T
	if n < len(slice)-1 {
		last = (slice)[n+1:]
	}
	tempSlice = first
	tempSlice = append(tempSlice, last...)
	return tempSlice
}

// bestimmt den absoluten Wert eines Ints
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
