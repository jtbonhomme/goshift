package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/fatih/color"
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

// User have a name, id, type, unavailable dates, and preferences.
type User struct {
	Name        string      `json:"name,omitempty"`
	Email       string      `json:"email,omitempty"`
	ID          string      `json:"id,omitempty"`
	Type        string      `json:"type,omitempty"`
	Unavailable []time.Time `json:"unavailable,omitempty"`
}

// Users lists all users.
type Users struct {
	Users []User `json:"users"`
}

// Override provides the start, end, user, and timezone of the override to work
// with the PagerDuty API.
type Override struct {
	Start time.Time    `json:"start"`
	End   time.Time    `json:"end"`
	User  AssignedUser `json:"user"`
}

// Overrides lists all overrides.
type Overrides struct {
	Overrides []Override `json:"overrides"`
}

// An AssignedUser has a name, id, and type for PagerDuty override.
type AssignedUser struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	ID    string `json:"id,omitempty"`
	Type  string `json:"type,omitempty"`
}

func (a AssignedUser) String() string {
	return a.Name
}

func parseFramadateCSV(data [][]string) Input {
	var dates []time.Time
	var input = Input{
		ScheduleStart: time.Now().Add(10 * 365 * OneDay),
		ScheduleEnd:   time.Now().Add(-10 * 365 * OneDay),
		Users:         []User{},
	}

	for i, line := range data {
		if i == 0 {
			//fmt.Println(line)
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

		//fmt.Println(line)
		user := User{
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

func main() {
	var err error

	fmt.Println("goshift")

	usersJson, err := os.Open("/Users/jean-thierry.bonhomme/Documents/Contentsquare/pagerduty-users.json")
	if err != nil {
	}
	fmt.Println("Successfully opened users.json")
	defer usersJson.Close()
	usersValue, _ := ioutil.ReadAll(usersJson)

	// we initialize our Users array
	var users Users

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(usersValue, &users)
	//fmt.Println(users)

	f, err := os.Open("/Users/jean-thierry.bonhomme/Downloads/DTOnCallFebruary2024.csv")
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

	primary, secondary, pstats, sstats, err := solver(input, users)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(primary.Overrides); i++ {
		weekday := primary.Overrides[i].Start.Weekday().String()
		margin := strings.Repeat(" ", 10-len(weekday))
		fmt.Printf("- %s %s %s: %s | %s\n", weekday, margin, primary.Overrides[i].Start, primary.Overrides[i].User, secondary.Overrides[i].User)
	}

	for i := 0; i < len(input.Users); i++ {
		fmt.Printf("* user %s: %d | %d\n", input.Users[i].Email, pstats[i], sstats[i])
	}

	p, err := json.MarshalIndent(primary, "", "  ")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("primary.json", p, 0644)
	if err != nil {
		panic(err)
	}

	displayCalendar("Primary on-call shift", primary)

	s, err := json.MarshalIndent(secondary, "", "  ")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("secondary.json", s, 0644)
	if err != nil {
		panic(err)
	}

	displayCalendar("Secondary on-call shift", secondary)
}

var months = [12]string{
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"August",
	"September",
	"October",
	"November",
	"December",
}

var days = [7]string{
	"Mon",
	"Tue",
	"Wed",
	"Thu",
	"Fri",
	"Sat",
	"Sun",
}

func displayCalendar(title string, schedule Overrides) {
	start := schedule.Overrides[0].Start
	month := start.Month()
	year := start.Year()

	fmt.Println(title)
	//fmt.Println(override.Overrides)
	printHeader(month, year)

	day := 1
	printFirstWeek(&day, start, schedule)
	printOtherWeeks(day, start, schedule)

	fmt.Println("")
}

const (
	monthStringLen string = "   "
	daySeparator   string = "         "
)

func printHeader(month time.Month, year int) {
	c1 := color.New(color.FgHiBlue).Add(color.Bold)
	c2 := color.New(color.FgHiCyan).Add(color.Bold)
	c1.Printf("%s %d", months[month-1], year)
	c1.Println("")
	c2.Println(strings.Join(days[:], daySeparator))
}

func printFirstWeek(day *int, start time.Time, schedule Overrides) {
	f := beginningOfMonth(start).Weekday()
	found := false

	for i, v := range days {
		if f.String()[0:3] == v {
			printDay(*day, i, start)
			*day++
			found = true
			continue
		} else {
			if !found {
				fmt.Print(daySeparator + monthStringLen)
			}
		}

		if found {
			printDay(*day, i, start)
			*day++
		}
	}

	fmt.Println("")
}

func printOtherWeeks(day int, start time.Time, schedule Overrides) {
	e := endOfMonth(start)
	idx := 0
	for day <= e.Day() {
		printDay(day, idx, start)
		idx++

		if idx >= len(days) {
			fmt.Println("")
			idx = 0
		}

		day++
	}
}

func printDay(day int, idx int, start time.Time) {
	workdayColor := color.New(color.FgWhite).Add(color.Bold)
	holidayColor := color.New(color.FgHiCyan).Add(color.Bold)
	currentDayColor := color.New(color.FgHiRed).Add(color.Bold)

	if day > 9 {
		if day == start.Day() {
			currentDayColor.Printf(" %d%s", day, daySeparator)
		} else if idx == 5 || idx == 6 {
			holidayColor.Printf(" %d%s", day, daySeparator)
		} else {
			workdayColor.Printf(" %d%s", day, daySeparator)
		}
	} else {
		if day == start.Day() {
			currentDayColor.Printf("  %d%s", day, daySeparator)
		} else if idx == 5 || idx == 6 {
			holidayColor.Printf("  %d%s", day, daySeparator)
		} else {
			workdayColor.Printf("  %d%s", day, daySeparator)
		}
	}
}

func beginningOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 0, -date.Day()+1)
}

func endOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 1, -date.Day())
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

func avg(stats []int) int {
	var cumul int
	l := len(stats)

	for _, s := range stats {
		cumul += s
	}
	return int(cumul / (l - 6))
}

func solver(input Input, users Users) (Overrides, Overrides, []int, []int, error) {
	var err error
	var overridesPrimary = Overrides{
		Overrides: []Override{},
	}
	var overridesSecondary = Overrides{
		Overrides: []Override{},
	}
	var primaryAvgShifts, secondaryAvgShifts int

	primaryStats := make([]int, len(input.Users))
	secondaryStats := make([]int, len(input.Users))

	ui := NewIterator(input.Users)

	// build shifts
	for d := input.ScheduleStart; d.Before(input.ScheduleEnd.Add(OneDay)); d = d.Add(OneDay) {
		weekday := d.Weekday().String()

		primary := Override{
			Start: d,
			End:   d.Add(OneDay),
		}

		// primary schedule override
		for i := 0; i < len(input.Users); i++ {
			user, n := ui.Next()
			if !slices.Contains(user.Unavailable, d) {
				// user not available on Sunday and current day is Saturday
				if weekday == time.Saturday.String() &&
					slices.Contains(user.Unavailable, d.Add(OneDay)) {
					continue
				}

				// already too much shifts for this user
				if primaryStats[n] > primaryAvgShifts {
					continue
				}

				u, err := retrieveUser(user, users)
				if err != nil {
					fmt.Printf("error: %s\n", err.Error())
					continue
				}

				primary.User = u
				primaryStats[n]++

				break
			}
		}

		secondary := Override{
			Start: d,
			End:   d.Add(OneDay),
		}

		// secondary schedule override
		for i := 0; i < len(input.Users); i++ {
			user, n := ui.Next()
			if !slices.Contains(user.Unavailable, d) {
				// no newbie as secondary at beginning
				if user.Email == "valerio.figliuolo@contentsquare.com" ||
					user.Email == "ahmed.khaled@contentsquare.com" ||
					user.Email == "houssem.touansi@contentsquare.com" ||
					user.Email == "kevin.albes@contentsquare.com" ||
					user.Email == "yunbo.wang@contentsquare.com" ||
					user.Email == "wael.tekaya@contentsquare.com" {
					continue
				}

				// user not available this day
				if weekday == time.Saturday.String() &&
					slices.Contains(user.Unavailable, d.Add(OneDay)) {
					continue
				}

				// already too much shifts for this user
				if secondaryStats[n] > secondaryAvgShifts {
					continue
				}

				u, err := retrieveUser(user, users)
				if err != nil {
					fmt.Printf("error: %s\n", err.Error())
					continue
				}

				secondary.User = u
				secondaryStats[n]++

				break
			}
		}

		// check shift
		if primary.User.Name == "" {
			// try to pick very first name available
			for i := 0; i < len(input.Users); i++ {
				user, n := ui.Next()
				if !slices.Contains(user.Unavailable, d) &&
					(weekday != time.Saturday.String() ||
						(weekday == time.Saturday.String() && !slices.Contains(user.Unavailable, d.Add(OneDay)))) {
					u, err := retrieveUser(user, users)
					if err != nil {
						fmt.Printf("error: %s\n", err.Error())
						continue
					}
					primary.User = u
					primaryStats[n]++
					break
				}
			}
			if primary.User.Name == "" {
				return Overrides{}, Overrides{}, nil, nil, fmt.Errorf("empty user for primary on %s", primary.Start)
			}
		}

		if secondary.User.Name == "" {
			// try to pick very first name available
			for i := 0; i < len(input.Users); i++ {
				user, n := ui.Next()
				if !slices.Contains(user.Unavailable, d) &&
					(weekday != time.Saturday.String() ||
						(weekday == time.Saturday.String() && !slices.Contains(user.Unavailable, d.Add(OneDay)))) {
					u, err := retrieveUser(user, users)
					if err != nil {
						fmt.Printf("error: %s\n", err.Error())
						continue
					}
					secondary.User = u
					secondaryStats[n]++
					break
				}
			}
			if secondary.User.Name == "" {
				return Overrides{}, Overrides{}, nil, nil, fmt.Errorf("empty user for secondary on %s", secondary.Start)
			}
		}

		if primary.User == secondary.User {
			return Overrides{}, Overrides{}, nil, nil, fmt.Errorf("same user for primary and secondary on %s", primary.Start)
		}

		overridesPrimary.Overrides = append(overridesPrimary.Overrides, primary)
		overridesSecondary.Overrides = append(overridesSecondary.Overrides, secondary)

		// weekday management
		if weekday == time.Saturday.String() && d.Before(input.ScheduleEnd) {
			overridesPrimary.Overrides = append(overridesPrimary.Overrides, Override{
				Start: primary.Start.Add(OneDay),
				End:   primary.End.Add(OneDay),
				User:  primary.User,
			})
			overridesSecondary.Overrides = append(overridesSecondary.Overrides, Override{
				Start: secondary.Start.Add(OneDay),
				End:   secondary.End.Add(OneDay),
				User:  secondary.User,
			})
			d = d.Add(OneDay)
		}

		primaryAvgShifts, secondaryAvgShifts = avg(primaryStats), avg(secondaryStats)
		//fmt.Printf("primaryAvgShifts %d\tsecondaryAvgShifts %d\n", primaryAvgShifts, secondaryAvgShifts)
	}

	return overridesPrimary, overridesSecondary, primaryStats, secondaryStats, err
}

func retrieveUser(user User, users Users) (AssignedUser, error) {
	for _, u := range users.Users {
		if u.Email == user.Email {
			return AssignedUser{
				Name:  u.Name,
				Email: u.Email,
				ID:    u.ID,
				Type:  u.Type,
			}, nil
		}
	}

	return AssignedUser{}, fmt.Errorf("unknown user %s", user.Email)
}
