package pagerduty

import "time"

type Unavalabilities struct {
	Weekdays map[string]int
	Weekends map[string]int
}

func (input *Input) UnavailablitiesStats() *Unavalabilities {
	unavailablitiesStats := &Unavalabilities{
		Weekdays: make(map[string]int),
		Weekends: make(map[string]int),
	}

	for _, user := range input.Users {
		for _, d := range user.Unavailable {
			if d.Weekday().String() != time.Saturday.String() {
				n := unavailablitiesStats.Weekdays[user.Email]
				unavailablitiesStats.Weekdays[user.Email] = n + 1
			} else if d.Weekday().String() == time.Saturday.String() {
				n := unavailablitiesStats.Weekends[user.Email]
				unavailablitiesStats.Weekends[user.Email] = n + 1
			}
		}
	}

	return unavailablitiesStats
}
