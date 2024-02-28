package solver

import (
	"fmt"
	"slices"
	"time"

	"github.com/jtbonhomme/goshift/internal/pagerduty"
	"github.com/jtbonhomme/goshift/internal/utils"
)

var newbies []string = []string{
	"valerio.figliuolo@contentsquare.com",
	"ahmed.khaled@contentsquare.com",
	"houssem.touansi@contentsquare.com",
	"kevin.albes@contentsquare.com",
	"yunbo.wang@contentsquare.com",
	"wael.tekaya@contentsquare.com",
}

func Run(input pagerduty.Input, users pagerduty.Users) (pagerduty.Overrides, pagerduty.Overrides, []int, []int, error) {
	var err error
	var overridesPrimary = pagerduty.Overrides{
		Overrides: []pagerduty.Override{},
	}
	var overridesSecondary = pagerduty.Overrides{
		Overrides: []pagerduty.Override{},
	}

	primaryStats := make([]int, len(input.Users))
	secondaryStats := make([]int, len(input.Users))

	ui := pagerduty.NewIterator(input.Users)

	// build shifts
	for d := input.ScheduleStart; d.Before(input.ScheduleEnd.Add(utils.OneDay)); d = d.Add(utils.OneDay) {
		weekday := d.Weekday().String()

		primary := pagerduty.Override{
			Start: d,
			End:   d.Add(utils.OneDay),
		}

		var nPrim, nSec int
		// primary schedule override
		for i := 0; i < len(input.Users); i++ {
			user, n := ui.Next()
			fmt.Printf("\t🅰️ [%s] considering %s (%d) with %d shifts for primary (avgShifts: %d - maxShifts: %d)", d.String(), user.Email, n, primaryStats[n], utils.Average(primaryStats, len(input.Users)), utils.Max(primaryStats))

			if !slices.Contains(user.Unavailable, d) {
				// user not available on Sunday and current day is Saturday
				if weekday == time.Saturday.String() &&
					slices.Contains(user.Unavailable, d.Add(utils.OneDay)) {
					fmt.Println(" not available on Sunday --> NEXT")
					continue
				}

				// already too much shifts for this user
				//if primaryStats[n] > utils.Average(primaryStats, len(input.Users)) {
				if primaryStats[n] > utils.Max(primaryStats) || primaryStats[n] > utils.MinWithoutZero(primaryStats)+1 || primaryStats[n] > utils.Average(primaryStats, len(input.Users))+1 {
					fmt.Println(" stats too high --> NEXT")
					continue
				}

				u, err := pagerduty.RetrieveUser(user, users)
				if err != nil {
					fmt.Printf("error: %s\n", err.Error())
					continue
				}

				primary.User = u
				primaryStats[n]++
				nPrim = n
				fmt.Println(" --> SELECTED")
				break
			}
			fmt.Println(" not available --> NEXT")
		}

		secondary := pagerduty.Override{
			Start: d,
			End:   d.Add(utils.OneDay),
		}

		// secondary schedule override
		for i := 0; i < len(input.Users); i++ {
			user, n := ui.Next()
			fmt.Printf("\t🅱️ [%s] considering %s (%d) with %d shifts for secondary (avgShifts: %d - maxShifts: %d)", d.String(), user.Email, n, primaryStats[n], utils.Average(secondaryStats, len(input.Users)), utils.Max(secondaryStats))

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
				if secondaryStats[n] > utils.MinWithoutZero(secondaryStats)+1 {
					fmt.Println(" stats too high --> NEXT")
					continue
				}

				u, err := pagerduty.RetrieveUser(user, users)
				if err != nil {
					fmt.Printf("error: %s\n", err.Error())
					continue
				}

				secondary.User = u
				secondaryStats[n]++
				nSec = n
				fmt.Printf(" --> SELECTED\n\n")
				break
			}
			fmt.Println(" not available --> NEXT")
		}

		// check shift
		if primary.User.Name == "" {
			fmt.Printf("\t⚠️ could not find any primary, need to reselect another user: ")
			// try to pick very first name available
			for i := 0; i < len(input.Users); i++ {
				user, n := ui.Next()
				if !slices.Contains(user.Unavailable, d) &&
					(weekday != time.Saturday.String() ||
						(weekday == time.Saturday.String() && !slices.Contains(user.Unavailable, d.Add(utils.OneDay)))) {
					u, err := pagerduty.RetrieveUser(user, users)
					if err != nil {
						fmt.Printf("error: %s\n", err.Error())
						continue
					}
					primary.User = u
					primaryStats[n]++
					nPrim = n
					fmt.Printf("%s\n", u.Name)
					break
				}
			}
			if primary.User.Name == "" {
				return pagerduty.Overrides{}, pagerduty.Overrides{}, nil, nil, fmt.Errorf("empty user for primary on %s", primary.Start)
			}
		}

		if secondary.User.Name == "" {
			// try to pick very first name available
			for i := 0; i < len(input.Users); i++ {
				user, n := ui.Next()
				if !slices.Contains(user.Unavailable, d) &&
					(weekday != time.Saturday.String() ||
						(weekday == time.Saturday.String() && !slices.Contains(user.Unavailable, d.Add(utils.OneDay)))) {
					// no newbie as secondary at beginning
					if slices.Contains(newbies, user.Email) {
						continue
					}

					u, err := pagerduty.RetrieveUser(user, users)
					if err != nil {
						fmt.Printf("error: %s\n", err.Error())
						continue
					}
					secondary.User = u
					secondaryStats[n]++
					nSec = n
					break
				}
			}
			if secondary.User.Name == "" {
				return pagerduty.Overrides{}, pagerduty.Overrides{}, nil, nil, fmt.Errorf("empty user for secondary on %s", secondary.Start)
			}
		}

		if primary.User == secondary.User {
			// try to pick very first other name available
			for i := 0; i < len(input.Users); i++ {
				user, n := ui.Next()
				if !slices.Contains(user.Unavailable, d) &&
					(weekday != time.Saturday.String() ||
						(weekday == time.Saturday.String() && !slices.Contains(user.Unavailable, d.Add(utils.OneDay)))) {
					// no newbie as secondary at beginning
					if slices.Contains(newbies, user.Email) {
						continue
					}

					u, err := pagerduty.RetrieveUser(user, users)
					if err != nil {
						fmt.Printf("error: %s\n", err.Error())
						continue
					}
					secondary.User = u
					secondaryStats[n]++
					nSec = n
					break
				}
			}

			if primary.User == secondary.User {
				return pagerduty.Overrides{}, pagerduty.Overrides{}, nil, nil, fmt.Errorf("same user for primary and secondary on %s", primary.Start)
			}
		}

		overridesPrimary.Overrides = append(overridesPrimary.Overrides, primary)
		overridesSecondary.Overrides = append(overridesSecondary.Overrides, secondary)

		// weekday management
		if weekday == time.Saturday.String() && d.Before(input.ScheduleEnd) {
			overridesPrimary.Overrides = append(overridesPrimary.Overrides, pagerduty.Override{
				Start: primary.Start.Add(utils.OneDay),
				End:   primary.End.Add(utils.OneDay),
				User:  primary.User,
			})
			overridesSecondary.Overrides = append(overridesSecondary.Overrides, pagerduty.Override{
				Start: secondary.Start.Add(utils.OneDay),
				End:   secondary.End.Add(utils.OneDay),
				User:  secondary.User,
			})
			primaryStats[nPrim]++
			secondaryStats[nSec]++
			d = d.Add(utils.OneDay)
		}
	}

	return overridesPrimary, overridesSecondary, primaryStats, secondaryStats, err
}
