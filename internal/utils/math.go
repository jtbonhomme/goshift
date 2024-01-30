package utils

func Average(stats []int, total int) int {
	var cumul int

	for _, s := range stats {
		cumul += s
	}

	return int(cumul / total)
}
