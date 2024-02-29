package solver

import (
	"time"

	"github.com/jtbonhomme/goshift/internal/pagerduty"
)

const (
	TopAvailabilityRation    = 4000
	BottomAvailabilityRation = 40
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

func sortAvailableUsers(users []pagerduty.User, d time.Time) []pagerduty.User {
	var sortedAvailableUsers = []pagerduty.User{}

	// rank users
	var rank = make([]int, len(users))
	for i, user := range users {
		rank[i] = int(TopAvailabilityRation / (BottomAvailabilityRation - len(user.Unavailable)))
	}

	// sort per ranking
	for i := 0; i < len(users); i++ {
		j := maxIndex(rank)
		sortedAvailableUsers = append(sortedAvailableUsers, users[j])
		rank[j] = -1
	}

	return sortedAvailableUsers
}
