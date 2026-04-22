package utils

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

var Logger *slog.Logger

// entfernt ntes Element aus slice
// ToDo error
func RemoveElementFromSlice[T any](slice []T, n int) ([]T, error) {
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

	return tempSlice, nil
}

// bestimmt den absoluten Wert eines Ints
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func Timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}

func CheckName(name string) error {
	if name == "" {
		return fmt.Errorf("Please provide a valid name.")
	}
	// ist es nur eine Nummer?
	_, err := strconv.Atoi(name)
	if err == nil {
		return fmt.Errorf("Please provide a name that is not only a number.")
	}
	return nil
}
