package utils

import (
	"time"

	"github.com/jtbonhomme/goshift/internal/pagerduty"
)

var location *time.Location

func init() { //nolint:gochecknoinits // todo
	var err error
	location, err = time.LoadLocation("Europe/Paris")
	if err != nil {
		panic(err)
	}
}

func ParseFramadateCSV(data [][]string) pagerduty.Input {
	var dates []time.Time
	var input = pagerduty.Input{
		ScheduleStart: time.Now().Add(10 * 365 * OneDay),
		ScheduleEnd:   time.Now().Add(-10 * 365 * OneDay),
		Users:         []pagerduty.User{},
	}

	for i, line := range data {
		if i == 0 {
			dates = make([]time.Time, len(line))
			for j, field := range line {
				if field == "" {
					continue
				}

				d, err := time.Parse("02/01/2006", field)
				if err != nil {
					panic(err)
				}

				// TZ
				t := time.Date(d.Year(), d.Month(), d.Day(), 9, 0, 0, 0, location)
				if input.ScheduleStart.After(d) {
					input.ScheduleStart = t
				}

				if input.ScheduleEnd.Before(d) {
					input.ScheduleEnd = t
				}

				dates[j] = t
			}
			continue
		}

		// empty line after header
		if i == 1 {
			continue
		}

		user := pagerduty.User{
			Unavailable: []time.Time{},
		}

		for j, field := range line {
			if j == 0 {
				user.Email = field
			} else if field == "Non" {
				user.Unavailable = append(user.Unavailable, dates[j])
			}
		}

		input.Users = append(input.Users, user)
	}

	return input
}
