package schedule

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/jtbonhomme/goshift/internal/pagerduty"
)

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

func DisplayCalendar(title string, schedule pagerduty.Overrides) {
	start := schedule.Overrides[0].Start
	month := start.Month()
	year := start.Year()

	fmt.Println(title)
	fmt.Println("")
	printHeader(month, year)

	day := 1
	printFirstWeek(&day, start, schedule)
	printOtherWeeks(day, start, schedule)

	fmt.Println("")
	fmt.Println("")
}

const (
	monthStringLen string = "   "
	daySeparator   string = " "
	nameLen        int    = 10
)

func printHeader(month time.Month, year int) {
	c1 := color.New(color.FgHiBlue).Add(color.Bold)
	c2 := color.New(color.FgHiCyan).Add(color.Bold)
	c1.Printf("%s %d", months[month-1], year)
	c1.Println("")
	c2.Println(strings.Join(days[:], daySeparator+strings.Repeat(" ", nameLen)))
}

func printFirstWeek(day *int, start time.Time, schedule pagerduty.Overrides) {
	f := beginningOfMonth(start).Weekday()
	found := false

	for i, v := range days {
		if f.String()[0:3] == v {
			printDay(*day, i, start)
			printName(*day, schedule)
			*day++
			found = true
			continue
		} else {
			if !found {
				fmt.Print(daySeparator + monthStringLen + strings.Repeat(" ", nameLen))
			}
		}

		if found {
			printDay(*day, i, start)
			printName(*day, schedule)
			*day++
		}
	}

	fmt.Println("")
}

func printOtherWeeks(day int, start time.Time, schedule pagerduty.Overrides) {
	e := endOfMonth(start)
	idx := 0
	for day <= e.Day() {
		printDay(day, idx, start)
		printName(day, schedule)
		idx++

		if idx >= len(days) {
			fmt.Println("")
			idx = 0
		}

		day++
	}
}

func printName(idx int, schedule pagerduty.Overrides) {
	firstName := strings.Split(schedule.Overrides[idx-1].User.Name, " ")[0]
	if len(firstName) > nameLen {
		firstName = firstName[:nameLen]
	} else {
		firstName += strings.Repeat(" ", nameLen-len(firstName))
	}

	fmt.Print(firstName)
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
