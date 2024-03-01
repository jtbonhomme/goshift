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

	// newbies are not allowed to do secondary
	if isSecondary {
		excludedUsers = s.newbies
	}

	// primary schedule override
	for i := 0; i < ui.Len(excludedUsers); i++ {
		user, _ := ui.NextWithExclude(excludedUsers)
		fmt.Printf("\t%s [%s] considering %s: %d | %d shifts (avgShifts: %d - avgWeekends: %d)", label, d.String(), user.Email, s.Stats[user.Email], s.WeekendStats[user.Email], utils.Average(s.Stats), utils.Average(s.WeekendStats))

		// already too much shifts for this user
		//if s.Stats[n] > utils.Max(stats) || s.Stats[n] > utils.MinWithoutZero(stats)+1 || s.Stats[n] > utils.Average(stats)+1 {
		if checkStats && s.Stats[user.Email] > utils.Average(s.Stats)+1 {
			fmt.Println(" stats too high --> NEXT")
			continue
		}

		if checkStats && weekday == time.Saturday.String() && s.WeekendStats[user.Email] > utils.Average(s.WeekendStats) {
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
		s.Stats[user.Email]++
		fmt.Println(" --> SELECTED")

		break
	}

	return override
}
