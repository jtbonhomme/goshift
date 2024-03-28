package solver

import (
	"slices"
	"time"

	"github.com/rs/zerolog/log"

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

	// schedule override
	for i := 0; i < len(s.input.Users); i++ {
		user, _, ok := ui.NextWithExclude(excludedUsers)
		// pick next user in the list
		if !ok {
			log.Debug().Msgf("\t%s [%s] error no result for next iterator with exclude", label, d.String())
			return override
		}

		log.Debug().Msgf("\t%s [%s] considering %s: %d | %d shifts (min Shifts: %d - min Weekends: %d | avg Shifts: %d - avg Weekends: %d)", label, d.String(), user.Email, s.Stats[user.Email], s.WeekendStats[user.Email], utils.Min(s.Stats), utils.Min(s.WeekendStats), utils.Average(s.Stats), utils.Average(s.WeekendStats))

		// if user is un available that day, move to the next user
		if slices.Contains(user.Unavailable, d) {
			log.Debug().Msg(" not available this day --> NEXT")
			continue
		}

		// already too much weekend shifts for this user
		if checkStats && weekday == time.Saturday.String() && s.WeekendStats[user.Email] > utils.Min(s.WeekendStats) {
			log.Debug().Msg(" too much week-ends (> min) --> NEXT")
			continue
		}

		// already too much week days shifts for this user
		if checkStats && weekday != time.Saturday.String() && s.Stats[user.Email] > utils.Min(s.Stats) {
			log.Debug().Msg(" stats too high (> min) --> NEXT")
			continue
		}

		u, err := s.users.RetrieveAssignedUser(user)
		if err != nil {
			log.Debug().Msgf("error: %s", err.Error())
			continue
		}

		if slices.Contains(lastUsers, u) {
			log.Debug().Msg(" user already selected previous day --> NEXT")
			continue
		}

		override.User = u
		s.Stats[user.Email]++
		log.Debug().Msg(" --> SELECTED")

		break
	}

	return override
}
