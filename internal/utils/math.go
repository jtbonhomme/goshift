package utils

import (
	"math"
)

func Average(stats []int, total int) int {
	var cumul int

	for _, s := range stats {
		cumul += s
	}

	return int(cumul / total)
}

func Max(stats []int) int {
	var max int

	for _, s := range stats {
		if s > max {
			max = s
		}
	}

	return max
}

func Min(stats []int) int {
	var min int

	min = math.MaxInt
	for _, s := range stats {
		if s < min {
			min = s
		}
	}

	return min
}

func MinWithoutZero(stats []int) int {
	var min int

	min = math.MaxInt
	for _, s := range stats {
		if s != 0 && s < min {
			min = s
		}
	}

	return min
}
