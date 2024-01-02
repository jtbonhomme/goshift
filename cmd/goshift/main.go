package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
)

const (
	OneDay time.Duration = time.Hour * 24
)

var location *time.Location

func init() {
	var err error
	location, err = time.LoadLocation("Europe/Paris")
	if err != nil {
		panic(err)
	}
}

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
	Start time.Time    `json:"start"`
	End   time.Time    `json:"end"`
	User  AssignedUser `json:"user"`
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

func parseFramadateCSV(data [][]string) Input {
	var dates []time.Time
	var input = Input{
		ScheduleStart: time.Now().Add(10*365 + 24*time.Hour),
		ScheduleEnd:   time.Now().Add(-10*365 + 24*time.Hour),
		Users:         []User{},
	}

	for i, line := range data {
		if i == 0 {
			fmt.Println(line)
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
		fmt.Println(line)
		user := User{
			Unavailable: []time.Time{},
		}
		for j, field := range line {
			if j == 0 {
				user.Name = field
			} else if field == "Non" {
				user.Unavailable = append(user.Unavailable, dates[j])
			}
		}
		input.Users = append(input.Users, user)
	}

	return input
}

func main() {
	var err error

	fmt.Println("goshift")

	f, err := os.Open("/Users/jbonhomm/Downloads/January2024DTOnCall.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}

	input := parseFramadateCSV(data)

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

	p, err := json.MarshalIndent(primary, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println("Primary on-call shift")
	fmt.Print(string(p))
	fmt.Println("")

	s, err := json.MarshalIndent(secondary, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println("Secondary on-call shift")
	fmt.Print(string(s))
	fmt.Println("")
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
		weekday := d.Weekday().String()

		primary := Override{
			Start: d,
			End:   d.Add(OneDay),
		}

		for i := 0; i < len(input.Users); i++ {
			user, n := ui.Next()
			if !slices.Contains(user.Unavailable, d) {
				if weekday == time.Saturday.String() &&
					slices.Contains(user.Unavailable, d.Add(OneDay)) {
					continue
				}
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
				if weekday == time.Saturday.String() &&
					slices.Contains(user.Unavailable, d.Add(OneDay)) {
					continue
				}
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
		if weekday == time.Saturday.String() && d.Before(input.ScheduleEnd) {
			overridesPrimary = append(overridesPrimary, Override{
				Start: primary.Start.Add(OneDay),
				End:   primary.End.Add(OneDay),
				User:  primary.User,
			})
			overridesSecondary = append(overridesSecondary, Override{
				Start: secondary.Start.Add(OneDay),
				End:   secondary.End.Add(OneDay),
				User:  secondary.User,
			})
			d = d.Add(OneDay)
		}
	}

	return overridesPrimary, overridesSecondary, primaryStats, secondaryStats, err
}
