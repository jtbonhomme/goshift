package solver

import (
	"math"

	"github.com/jtbonhomme/goshift/internal/pagerduty"
)

const (
	TopAvailabilityRatio    int = 250
	BottomAvailabilityRatio int = 50
	StatsPenaltyFactor      int = 2
)

func maxIndex(a []int) int {
	var max, maxIndex int

	for i := 0; i < len(a); i++ {
		if a[i] > max {
			max = a[i]
			maxIndex = i
		}
	}

	return maxIndex
}

func minIndex(a []int) int {
	var min, minIndex int
	min = math.MaxInt

	for i := 0; i < len(a); i++ {
		if a[i] < min {
			min = a[i]
			minIndex = i
		}
	}

	return minIndex
}

func sortUsersPerAvailability(users []pagerduty.User) []pagerduty.User {
	var sortedUsers = []pagerduty.User{}

	// rank users
	var rank = make([]int, len(users))
	for i, user := range users {
		rank[i] = int(TopAvailabilityRatio / (BottomAvailabilityRatio - len(user.Unavailable)))
	}

	// sort per ranking
	for i := 0; i < len(users); i++ {
		j := maxIndex(rank)
		sortedUsers = append(sortedUsers, users[j])
		rank[j] = -1
	}

	return sortedUsers
}

func sortUsersPerStats(users []pagerduty.User, stats map[string]int) []pagerduty.User {
	var sortedUsers = []pagerduty.User{}

	// rank users
	var rank = make([]int, len(users))
	for i, user := range users {
		rank[i] = stats[user.Email]
	}

	// sort per ranking
	for i := 0; i < len(users); i++ {
		j := minIndex(rank)
		sortedUsers = append(sortedUsers, users[j])
		rank[j] = math.MaxInt
	}

	return sortedUsers
}

func sortUsersPerAvailabilityAndStats(users []pagerduty.User, stats map[string]int) []pagerduty.User {
	var sortedUsers = []pagerduty.User{}

	// rank users
	var rank = make([]int, len(users))
	for i, user := range users {
		r := int(TopAvailabilityRatio/(BottomAvailabilityRatio-len(user.Unavailable))) - stats[user.Email]*StatsPenaltyFactor
		if r <= 0 {
			r = 1
		}
		rank[i] = r
	}

	// sort per ranking
	for i := 0; i < len(users); i++ {
		j := maxIndex(rank)
		sortedUsers = append(sortedUsers, users[j])
		rank[j] = -1
	}

	return sortedUsers
}
