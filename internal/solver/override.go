package solver

import (
	"fmt"
	"slices"
	"time"

	"github.com/jtbonhomme/goshift/internal/pagerduty"
	"github.com/jtbonhomme/goshift/internal/utils"
)

func (s *Solver) nextAvailableUser(d time.Time) (pagerduty.User, int, bool) {
	var found bool
	var n int
	var user pagerduty.User
	weekday := d.Weekday().String()

	for i := 0; i < len(s.input.Users) && !found; i++ {
		user, n = s.ui.Next()
		if !slices.Contains(user.Unavailable, d) {
			// user not available on Sunday and current day is Saturday
			if weekday == time.Saturday.String() &&
				slices.Contains(user.Unavailable, d.Add(utils.OneDay)) {
				fmt.Printf("\t\t %s not available on Sunday --> NEXT\n", user.Name)
				continue
			}
			found = true
			break
		}
	}

	return user, n, found
}

func (s *Solver) processPrimaryOverride(d time.Time, lastPrimaryUser pagerduty.AssignedUser) (pagerduty.Override, int) {
	var nPrim int
	primary := pagerduty.Override{
		Start: d,
		End:   d.Add(utils.OneDay),
	}
	weekday := d.Weekday().String()

	// primary schedule override
	for i := 0; i < len(s.input.Users); i++ {
		user, n := s.ui.Next()
		fmt.Printf("\t🅰️ [%s] considering %s (%d) with %d shifts for primary (avgShifts: %d - maxShifts: %d)", d.String(), user.Email, n, s.primaryStats[n], utils.Average(s.primaryStats, len(s.input.Users)), utils.Max(s.primaryStats))

		if !slices.Contains(user.Unavailable, d) {
			// user not available on Sunday and current day is Saturday
			if weekday == time.Saturday.String() &&
				slices.Contains(user.Unavailable, d.Add(utils.OneDay)) {
				fmt.Println(" not available on Sunday --> NEXT")
				continue
			}

			// already too much shifts for this user
			if s.primaryStats[n] > utils.Max(s.primaryStats) || s.primaryStats[n] > utils.MinWithoutZero(s.primaryStats)+1 || s.primaryStats[n] > utils.Average(s.primaryStats, len(s.input.Users))+1 {
				fmt.Println(" stats too high --> NEXT")
				continue
			}

			u, err := pagerduty.RetrieveUser(user, s.users)
			if err != nil {
				fmt.Printf("error: %s\n", err.Error())
				continue
			}

			if u == lastPrimaryUser {
				fmt.Println(" user already selected previous day --> NEXT")
				continue
			}

			if weekday == time.Saturday.String() && s.weekendStats[n] > utils.Average(s.weekendStats, len(s.input.Users)) {
				fmt.Println(" too much week-ends --> NEXT")
				continue
			}

			primary.User = u
			s.primaryStats[n]++
			nPrim = n
			fmt.Println(" --> SELECTED")

			break
		}
		fmt.Println(" not available --> NEXT")
	}

	return primary, nPrim
}

func (s *Solver) processSecondaryOverride(d time.Time, lastSecondaryUser pagerduty.AssignedUser) (pagerduty.Override, int) {
	var nSec int
	secondary := pagerduty.Override{
		Start: d,
		End:   d.Add(utils.OneDay),
	}
	weekday := d.Weekday().String()

	// secondary schedule override
	for i := 0; i < len(s.input.Users); i++ {
		user, n := s.ui.Next()
		fmt.Printf("\t🅱️ [%s] considering %s (%d) with %d shifts for secondary (avgShifts: %d - maxShifts: %d)", d.String(), user.Email, n, s.primaryStats[n], utils.Average(s.secondaryStats, len(s.input.Users)), utils.Max(s.secondaryStats))

		if !slices.Contains(user.Unavailable, d) {
			// no newbie as secondary at beginning
			if slices.Contains(newbies, user.Email) {
				fmt.Println(" is a newbie --> NEXT")
				continue
			}

			// user not available this day
			if weekday == time.Saturday.String() &&
				slices.Contains(user.Unavailable, d.Add(utils.OneDay)) {
				fmt.Println(" not available on Sunday --> NEXT")
				continue
			}

			// already too much shifts for this user
			if s.secondaryStats[n] > utils.MinWithoutZero(s.secondaryStats)+1 {
				fmt.Println(" stats too high --> NEXT")
				continue
			}

			u, err := pagerduty.RetrieveUser(user, s.users)
			if err != nil {
				fmt.Printf("error: %s\n", err.Error())
				continue
			}

			if u == lastSecondaryUser {
				fmt.Println(" user already selected previous day --> NEXT")
				continue
			}

			if weekday == time.Saturday.String() && s.weekendStats[n] > utils.Average(s.weekendStats, len(s.input.Users)) {
				fmt.Println(" too much week-ends --> NEXT")
				continue
			}

			secondary.User = u
			s.secondaryStats[n]++
			nSec = n
			fmt.Printf(" --> SELECTED\n\n")
			break
		}
		fmt.Println(" not available --> NEXT")
	}

	return secondary, nSec
}
