package util

import "math/rand"

func StringInSlice(arr []string, val string) bool {
	for _, s := range arr {
		if s == val {
			return true
		}
	}

	return false
}

func IntInSlice(arr []int, val int) bool {
	for _, s := range arr {
		if s == val {
			return true
		}
	}

	return false
}

// GetRandomNumberInRange returns a random number between
// min (inclusive) and max (exclusive).
func GetRandomNumberInRange(min, max int) int {
	if min == max {
		return min
	}

	return min + rand.Intn(max-min)
}
