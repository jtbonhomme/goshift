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

type Solver struct {
	input          pagerduty.Input
	users          pagerduty.Users
	primaryStats   []int
	weekendStats   []int
	secondaryStats []int
	ui             *pagerduty.UserIterator
}

func New(input pagerduty.Input, users pagerduty.Users) *Solver {
	primaryStats := make([]int, len(input.Users))
	weekendStats := make([]int, len(input.Users))
	secondaryStats := make([]int, len(input.Users))
	ui := pagerduty.NewIterator(input.Users)

	return &Solver{
		input:          input,
		users:          users,
		primaryStats:   primaryStats,
		weekendStats:   weekendStats,
		secondaryStats: secondaryStats,
		ui:             ui,
	}
}

func (s *Solver) Run() (pagerduty.Overrides, pagerduty.Overrides, []int, []int, []int, error) {
	var err error
	var overridesPrimary = pagerduty.Overrides{
		Overrides: []pagerduty.Override{},
	}
	var overridesSecondary = pagerduty.Overrides{
		Overrides: []pagerduty.Override{},
	}

	var lastPrimaryUser, lastSecondaryUser pagerduty.AssignedUser

	// build shifts
	for d := s.input.ScheduleStart; d.Before(s.input.ScheduleEnd.Add(utils.OneDay)); d = d.Add(utils.OneDay) {
		weekday := d.Weekday().String()

		primary, nPrim := s.processPrimaryOverride(d, lastPrimaryUser)
		lastPrimaryUser = primary.User
		secondary, nSec := s.processSecondaryOverride(d, lastSecondaryUser)
		lastSecondaryUser = secondary.User

		// check shift
		if primary.User.Name == "" {
			fmt.Printf("\t⚠️ could not find any primary, need to reselect another user: ")
			// try to pick very first name available
			for i := 0; i < len(s.input.Users); i++ {
				user, n := s.ui.Next()
				if !slices.Contains(user.Unavailable, d) &&
					(weekday != time.Saturday.String() ||
						(weekday == time.Saturday.String() && !slices.Contains(user.Unavailable, d.Add(utils.OneDay)))) {
					u, err := pagerduty.RetrieveUser(user, s.users)
					if err != nil {
						fmt.Printf("error: %s\n", err.Error())
						continue
					}
					primary.User = u
					s.primaryStats[n]++
					nPrim = n
					fmt.Printf("%s\n", u.Name)
					break
				}
			}
			if primary.User.Name == "" {
				return pagerduty.Overrides{}, pagerduty.Overrides{}, nil, nil, nil, fmt.Errorf("empty user for primary on %s", primary.Start)
			}
		}

		if secondary.User.Name == "" {
			// try to pick very first name available
			for i := 0; i < len(s.input.Users); i++ {
				user, n := s.ui.Next()
				if !slices.Contains(user.Unavailable, d) &&
					(weekday != time.Saturday.String() ||
						(weekday == time.Saturday.String() && !slices.Contains(user.Unavailable, d.Add(utils.OneDay)))) {
					// no newbie as secondary at beginning
					if slices.Contains(newbies, user.Email) {
						continue
					}

					u, err := pagerduty.RetrieveUser(user, s.users)
					if err != nil {
						fmt.Printf("error: %s\n", err.Error())
						continue
					}
					secondary.User = u
					s.secondaryStats[n]++
					nSec = n
					break
				}
			}
			if secondary.User.Name == "" {
				return pagerduty.Overrides{}, pagerduty.Overrides{}, nil, nil, nil, fmt.Errorf("empty user for secondary on %s", secondary.Start)
			}
		}

		if primary.User == secondary.User {
			// try to pick very first other name available
			for i := 0; i < len(s.input.Users); i++ {
				user, n := s.ui.Next()
				if !slices.Contains(user.Unavailable, d) &&
					(weekday != time.Saturday.String() ||
						(weekday == time.Saturday.String() && !slices.Contains(user.Unavailable, d.Add(utils.OneDay)))) {
					// no newbie as secondary at beginning
					if slices.Contains(newbies, user.Email) {
						continue
					}

					u, err := pagerduty.RetrieveUser(user, s.users)
					if err != nil {
						fmt.Printf("error: %s\n", err.Error())
						continue
					}
					secondary.User = u
					s.secondaryStats[n]++
					nSec = n
					break
				}
			}

			if primary.User == secondary.User {
				return pagerduty.Overrides{}, pagerduty.Overrides{}, nil, nil, nil, fmt.Errorf("same user for primary and secondary on %s", primary.Start)
			}
		}

		overridesPrimary.Overrides = append(overridesPrimary.Overrides, primary)
		overridesSecondary.Overrides = append(overridesSecondary.Overrides, secondary)

		// weekday management
		if weekday == time.Saturday.String() && d.Before(s.input.ScheduleEnd) {
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
			s.primaryStats[nPrim]++
			s.secondaryStats[nSec]++
			s.weekendStats[nPrim]++
			s.weekendStats[nSec]++
			d = d.Add(utils.OneDay)
		}
	}

	return overridesPrimary, overridesSecondary, s.primaryStats, s.secondaryStats, s.weekendStats, err
}
