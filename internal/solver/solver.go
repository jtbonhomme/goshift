package solver

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/jtbonhomme/goshift/internal/pagerduty"
	"github.com/jtbonhomme/goshift/internal/utils"
)

type Solver struct {
	input             pagerduty.Input
	users             pagerduty.Users
	Stats             map[string]int
	WeekendStats      map[string]int
	newbies           []string
	lastAssignedUsers []pagerduty.AssignedUser
}

func New(input pagerduty.Input, users pagerduty.Users, newbies, lastUsers []string) *Solver {
	// initialize maps
	Stats := make(map[string]int, len(input.Users))
	WeekendStats := make(map[string]int, len(input.Users))

	for _, user := range input.Users {
		Stats[user.Email] = 0
		WeekendStats[user.Email] = 0
	}

	lastAssignedUsers := []pagerduty.AssignedUser{}
	for _, email := range lastUsers {
		u, err := users.RetrieveAssignedUserByEmail(email)
		if err != nil {
			log.Debug().Msgf("error: %s", err.Error())
			continue
		}
		lastAssignedUsers = append(lastAssignedUsers, u)
	}

	return &Solver{
		input:             input,
		users:             users,
		Stats:             Stats,
		WeekendStats:      WeekendStats,
		newbies:           newbies,
		lastAssignedUsers: lastAssignedUsers,
	}

}

func (s *Solver) Run() (pagerduty.Overrides, pagerduty.Overrides, error) {
	var err error
	var overridesPrimary = pagerduty.Overrides{
		Overrides: []pagerduty.Override{},
	}
	var overridesSecondary = pagerduty.Overrides{
		Overrides: []pagerduty.Override{},
	}

	// build shifts
	for d := s.input.ScheduleStart; d.Before(s.input.ScheduleEnd.Add(utils.OneDay)); d = d.Add(utils.OneDay) {
		weekday := d.Weekday().String()

		// rank and sort available users depending of their number of available days
		sortedUsers := sortUsers(s.input.Users, s.Stats, "PerAvailabilitySimple")
		ui := pagerduty.NewIterator(sortedUsers)

		primary := s.processOverride("🅰️", d, s.lastAssignedUsers, ui, false, true)
		s.lastAssignedUsers = append(s.lastAssignedUsers, primary.User)
		secondary := s.processOverride("🅱️", d, s.lastAssignedUsers, ui, true, true)

		// check shift
		if primary.User.Name == "" {
			log.Debug().Msg("⚠️ \tcould not find any primary, need to reselect another user \t⚠️")
			// rank and sort available users depending of their stats
			sorted := sortUsers(s.input.Users, s.Stats, "PerStats")
			sui := pagerduty.NewIterator(sorted)
			// try to pick very first name available
			primary = s.processOverride("🅰️", d, s.lastAssignedUsers, sui, false, false)
			s.lastAssignedUsers = append(s.lastAssignedUsers, primary.User)
			if primary.User.Name == "" {
				return pagerduty.Overrides{}, pagerduty.Overrides{}, fmt.Errorf("empty user for primary on %s", primary.Start)
			}
		}

		if secondary.User.Name == "" {
			log.Debug().Msg("⚠️ \tcould not find any secondary, need to reselect another user \t⚠️")
			// rank and sort available users depending of their stats
			sorted := sortUsers(s.input.Users, s.Stats, "PerStats")
			sui := pagerduty.NewIterator(sorted)
			// try to pick very first name available
			secondary = s.processOverride("🅱️", d, s.lastAssignedUsers, sui, true, false)
			s.lastAssignedUsers = append(s.lastAssignedUsers, secondary.User)
			if secondary.User.Name == "" {
				return pagerduty.Overrides{}, pagerduty.Overrides{}, fmt.Errorf("empty user for secondary on %s", secondary.Start)
			}
		}

		if primary.User == secondary.User {
			return pagerduty.Overrides{}, pagerduty.Overrides{}, fmt.Errorf("same user for primary and secondary on %s", primary.Start)
		}

		log.Debug().Msg("")

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

			s.Stats[primary.User.Email]++
			s.Stats[secondary.User.Email]++
			s.WeekendStats[primary.User.Email]++
			s.WeekendStats[secondary.User.Email]++
			d = d.Add(utils.OneDay)
		}

		s.lastAssignedUsers = []pagerduty.AssignedUser{}
		s.lastAssignedUsers = append(s.lastAssignedUsers, primary.User)
		s.lastAssignedUsers = append(s.lastAssignedUsers, secondary.User)
	}

	return overridesPrimary, overridesSecondary, err
}
