package solver

import (
	"slices"
	"time"

	"github.com/jtbonhomme/goshift/internal/pagerduty"
	"github.com/jtbonhomme/goshift/internal/utils"
)

func filterUnavailableUsers(users []pagerduty.User, d time.Time) []pagerduty.User {
	var availableUsersOnly = []pagerduty.User{}
	weekday := d.Weekday().String()

	for _, user := range users {
		if (slices.Contains(user.Unavailable, d)) ||
			(!slices.Contains(user.Unavailable, d) &&
				weekday == time.Saturday.String() &&
				slices.Contains(user.Unavailable, d.Add(utils.OneDay))) {
			continue
		}
		availableUsersOnly = append(availableUsersOnly, user)
	}

	return availableUsersOnly
}
