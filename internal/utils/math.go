package utils

func Average(stats []int) int {
	var cumul int
	l := len(stats)

	for _, s := range stats {
		cumul += s
	}
	return int(cumul / (l - 6))
}
