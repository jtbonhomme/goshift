package utils

import (
	"math"
)

func Average(stats map[string]int) int {
	var cumul int

	for _, val := range stats {
		cumul += val
	}

	return cumul / len(stats)
}

func Max(stats map[string]int) int {
	var max int

	for _, val := range stats {
		if val > max {
			max = val
		}
	}

	return max
}

func Min(stats map[string]int) int {
	var min int

	min = math.MaxInt
	for _, val := range stats {
		if val < min {
			min = val
		}
	}

	return min
}

func MinWithoutZero(stats map[string]int) int {
	var min int

	min = math.MaxInt
	for _, val := range stats {
		if val != 0 && val < min {
			min = val
		}
	}

	return min
}
