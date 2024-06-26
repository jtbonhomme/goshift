package solver

import (
	"math"
	"time"

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
		rank[i] = TopAvailabilityRatio / (BottomAvailabilityRatio - len(user.Unavailable))
	}

	// sort per ranking
	for i := 0; i < len(users); i++ {
		j := maxIndex(rank)
		sortedUsers = append(sortedUsers, users[j])
		rank[j] = -1
	}

	return sortedUsers
}

func sortUsersPerAvailabilitySimple(users []pagerduty.User) []pagerduty.User {
	var sortedUsers = []pagerduty.User{}

	// rank users
	var rank = make([]int, len(users))
	for i, user := range users {
		rank[i] = len(user.Unavailable)
	}

	// sort per ranking
	for i := 0; i < len(users); i++ {
		j := maxIndex(rank)
		sortedUsers = append(sortedUsers, users[j])
		rank[j] = -1
	}

	return sortedUsers
}

func sortUsersPerRemainingAvailability(d time.Time, users []pagerduty.User) []pagerduty.User {
	var sortedUsers = []pagerduty.User{}

	// rank users
	var rank = make([]int, len(users))
	for i, user := range users {
		for _, a := range user.Unavailable {
			if a.After(d) &&
				(d.Weekday().String() != time.Saturday.String() || (d.Weekday().String() == time.Saturday.String() &&
					a.Weekday().String() == time.Saturday.String())) {
				rank[i]++
			}
		}
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
		r := TopAvailabilityRatio/(BottomAvailabilityRatio-len(user.Unavailable)) - stats[user.Email]*StatsPenaltyFactor
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

func sortUsers(d time.Time, users []pagerduty.User, stats map[string]int, method string) []pagerduty.User {
	switch method {
	case "PerAvailabilitySimple":
		return sortUsersPerAvailabilitySimple(users)
	case "PerRemainingAvailability":
		return sortUsersPerRemainingAvailability(d, users)
	case "PerAvailability":
		return sortUsersPerAvailability(users)
	case "PerStats":
		return sortUsersPerStats(users, stats)
	case "PerAvailabilityAndStats":
		return sortUsersPerAvailabilityAndStats(users, stats)
	default:
		return sortUsersPerAvailabilityAndStats(users, stats)
	}
}
