package solver

import (
	"fmt"
	"slices"
	"time"

	"github.com/jtbonhomme/goshift/internal/pagerduty"
	"github.com/jtbonhomme/goshift/internal/utils"
)

func (s *Solver) processOverride(label string, d time.Time, lastUsers []pagerduty.AssignedUser, ui *pagerduty.UserIterator, isSecondary, checkStats bool) pagerduty.Override {
	override := pagerduty.Override{
		Start: d,
		End:   d.Add(utils.OneDay),
	}

	weekday := d.Weekday().String()

	var excludedUsers = []string{}
	var stats = s.PrimaryStats
	if isSecondary {
		excludedUsers = s.secondaryExcludedUsers
		stats = s.SecondaryStats
	}

	// primary schedule override
	for i := 0; i < ui.Len(excludedUsers); i++ {
		user, _ := ui.NextWithExclude(excludedUsers)
		fmt.Printf("\t%s [%s] considering %s: %d | %d shifts (avgShifts: %d - maxShifts: %d)", label, d.String(), user.Email, stats[user.Email], s.WeekendStats[user.Email], utils.Average(stats, ui.Len(excludedUsers)), utils.Max(stats))

		// already too much shifts for this user
		//if stats[n] > utils.Max(stats) || stats[n] > utils.MinWithoutZero(stats)+1 || stats[n] > utils.Average(stats, ui.Len(excludedUsers))+1 {
		if checkStats && stats[user.Email] > utils.Average(stats, ui.Len(excludedUsers))+1 {
			fmt.Println(" stats too high --> NEXT")
			continue
		}

		if checkStats && weekday == time.Saturday.String() && s.WeekendStats[user.Email] > utils.Average(s.WeekendStats, ui.Len(excludedUsers)) {
			fmt.Println(" too much week-ends --> NEXT")
			continue
		}

		u, err := pagerduty.RetrieveAssignedUser(user, s.users)
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			continue
		}

		if slices.Contains(lastUsers, u) {
			fmt.Println(" user already selected previous day --> NEXT")
			continue
		}

		override.User = u
		if isSecondary {
			n := s.SecondaryStats[user.Email]
			s.SecondaryStats[user.Email] = n + 1
		} else {
			n := s.PrimaryStats[user.Email]
			s.PrimaryStats[user.Email] = n + 1
		}
		fmt.Println(" --> SELECTED")

		break
	}

	return override
}
