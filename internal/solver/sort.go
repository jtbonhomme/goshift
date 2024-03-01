package solver

import (
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

func sortUsers(users []pagerduty.User) []pagerduty.User {
	var sortedUsers = []pagerduty.User{}

	// rank users
	var rank = make([]int, len(users))
	for i, user := range users {
		rank[i] = int(TopAvailabilityRation / (BottomAvailabilityRation - len(user.Unavailable)))
	}

	// sort per ranking
	for i := 0; i < len(users); i++ {
		j := maxIndex(rank)
		sortedUsers = append(sortedUsers, users[j])
		rank[j] = -1
	}

	return sortedUsers
}
