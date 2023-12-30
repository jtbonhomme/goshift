package main

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"
)

const (
	OneDay time.Duration = time.Hour * 24
)

// Input for the pager duty scheduling problem. We have
// pager duty users that need to be assigned to days between the schedule start
// date and the schedule end date.
type Input struct {
	ScheduleStart time.Time `json:"schedule_start"`
	ScheduleEnd   time.Time `json:"schedule_end"`
	Users         []User    `json:"users"`
}

// Users have a name, id, type, unavailable dates, and preferences.
type User struct {
	Name        string      `json:"name,omitempty"`
	ID          string      `json:"id,omitempty"`
	Type        string      `json:"type,omitempty"`
	Unavailable []time.Time `json:"unavailable,omitempty"`
}

// Override provides the start, end, user, and timezone of the override to work
// with the PagerDuty API.
type Override struct {
	Start    time.Time    `json:"start"`
	End      time.Time    `json:"end"`
	User     AssignedUser `json:"user"`
	TimeZone string       `json:"time_zone"`
}

// An AssignedUser has a name, id, and type for PagerDuty override.
type AssignedUser struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

func (a AssignedUser) String() string {
	return a.Name
}

func main() {
	fmt.Println("goshift")
	inputJSON := `{
		"schedule_start": "2024-01-01T09:00:00+01:00",
		"schedule_end": "2024-02-01T09:00:00+01:00",
		"users": [
		  {
			"name": "a",
			"unavailable": [
				"2024-01-01T09:00:00+01:00"
			]
		  },
		  {
			"name": "t",
			"unavailable": [
			  "2024-01-14T09:00:00+01:00",
			  "2024-01-15T09:00:00+01:00",
			  "2024-01-16T09:00:00+01:00",
			  "2024-01-17T09:00:00+01:00",
			  "2024-01-18T09:00:00+01:00",
			  "2024-01-19T09:00:00+01:00",
			  "2024-01-20T09:00:00+01:00",
			  "2024-01-21T09:00:00+01:00",
			  "2024-01-22T09:00:00+01:00",
			  "2024-01-23T09:00:00+01:00",
			  "2024-01-24T09:00:00+01:00",
			  "2024-01-25T09:00:00+01:00",
			  "2024-01-26T09:00:00+01:00",
			  "2024-01-27T09:00:00+01:00",
			  "2024-01-28T09:00:00+01:00",
			  "2024-01-29T09:00:00+01:00",
			  "2024-01-30T09:00:00+01:00",
			  "2024-01-31T09:00:00+01:00"
			]
		  },
		  {
			"name": "c",
			"unavailable": [
				"2024-01-01T09:00:00+01:00"
			]
		  },
		  {
			"name": "e",
			"unavailable": [
				"2024-01-19T09:00:00+01:00",
				"2024-01-20T09:00:00+01:00",
				"2024-01-21T09:00:00+01:00"
			]
		  },
		  {
			"name": "o",
			"unavailable": [
				"2024-01-01T09:00:00+01:00",
				"2024-01-02T09:00:00+01:00",
				"2024-01-03T09:00:00+01:00",
				"2024-01-04T09:00:00+01:00",
				"2024-01-05T09:00:00+01:00",
				"2024-01-06T09:00:00+01:00",
				"2024-01-07T09:00:00+01:00",
				"2024-01-08T09:00:00+01:00",
				"2024-01-09T09:00:00+01:00",
				"2024-01-11T09:00:00+01:00"
			]
		  },
		  {
			"name": "y",
			"unavailable": [
				"2024-01-01T09:00:00+01:00",
				"2024-01-06T09:00:00+01:00",
				"2024-01-13T09:00:00+01:00",
				"2024-01-20T09:00:00+01:00",
				"2024-01-27T09:00:00+01:00"
			]
		  }
		 ]
	   }`

	input := Input{}

	err := json.Unmarshal([]byte(inputJSON), &input)
	if err != nil {
		panic(err)
	}

	primary, secondary, pstats, sstats, err := solver(input)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(primary); i++ {
		weekday := primary[i].Start.Weekday().String()
		margin := strings.Repeat(" ", 10-len(weekday))
		fmt.Printf("- %s %s %s: %s | %s\n", weekday, margin, primary[i].Start, primary[i].User, secondary[i].User)
	}

	for i := 0; i < len(input.Users); i++ {
		fmt.Printf("* user %s: %d | %d\n", input.Users[i].Name, pstats[i], sstats[i])
	}
}

type UserIterator struct {
	Users    []User
	iterator int
}

func NewIterator(users []User) *UserIterator {
	ui := UserIterator{
		Users: users,
	}

	return &ui
}

func (ui *UserIterator) Next() (User, int) {
	k := ui.iterator % len(ui.Users)
	ui.iterator++
	return ui.Users[k], k
}

func solver(input Input) ([]Override, []Override, []int, []int, error) {
	var err error
	var overridesPrimary = []Override{}
	var overridesSecondary = []Override{}
	primaryStats := make([]int, len(input.Users))
	secondaryStats := make([]int, len(input.Users))

	ui := NewIterator(input.Users)

	// build shifts
	for d := input.ScheduleStart; d.Before(input.ScheduleEnd); d = d.Add(OneDay) {
		primary := Override{
			Start: d,
			End:   d.Add(OneDay),
		}

		for i := 0; i < len(input.Users); i++ {
			user, n := ui.Next()
			if !slices.Contains(user.Unavailable, d) {
				primary.User = AssignedUser{
					Name: user.Name,
					ID:   user.ID,
					Type: user.Type,
				}
				primaryStats[n]++
				break
			}
		}

		secondary := Override{
			Start: d,
			End:   d.Add(OneDay),
		}

		for i := 0; i < len(input.Users); i++ {
			user, n := ui.Next()
			if !slices.Contains(user.Unavailable, d) {
				secondary.User = AssignedUser{
					Name: user.Name,
					ID:   user.ID,
					Type: user.Type,
				}
				secondaryStats[n]++
				break
			}
		}

		// check shift
		if primary.User.Name == "" {
			return nil, nil, nil, nil, fmt.Errorf("empty user for primary on %s", primary.Start)
		}

		if secondary.User.Name == "" {
			return nil, nil, nil, nil, fmt.Errorf("empty user for secondary on %s", secondary.Start)
		}

		if primary.User == secondary.User {
			return nil, nil, nil, nil, fmt.Errorf("same user for primary and secondary on %s", primary.Start)
		}

		overridesPrimary = append(overridesPrimary, primary)
		overridesSecondary = append(overridesSecondary, secondary)
	}

	return overridesPrimary, overridesSecondary, primaryStats, secondaryStats, err
}
