package schedule

import (
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rs/zerolog/log"

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

	log.Info().Msg(title)
	log.Info().Msg("")
	printHeader(month, year)

	day := 1
	printFirstWeek(&day, start, schedule)
	printOtherWeeks(day, start, schedule)

	log.Info().Msg("")
	log.Info().Msg("")
}

const (
	monthStringLen string = "   "
	daySeparator   string = " "
	nameLen        int    = 10
)

func printHeader(month time.Month, year int) {
	c1 := color.New(color.FgHiBlue).Add(color.Bold)
	c2 := color.New(color.FgHiCyan).Add(color.Bold)
	log.Info().Msg(c1.Sprintf("%s %d", months[month-1], year))
	log.Info().Msg("")
	log.Info().Msg(c2.Sprint(strings.Join(days[:], daySeparator+strings.Repeat(" ", nameLen))))
}

func printFirstWeek(day *int, start time.Time, schedule pagerduty.Overrides) {
	f := beginningOfMonth(start).Weekday()
	found := false

	line := ""
	for i, v := range days {
		if f.String()[0:3] == v {
			line += printDay(*day, i, start)
			line += printName(*day, schedule)
			*day++
			found = true
			continue
		} else if !found {
			line += daySeparator + monthStringLen + strings.Repeat(" ", nameLen)
		}

		if found {
			line += printDay(*day, i, start)
			line += printName(*day, schedule)
			*day++
		}
	}

	log.Info().Msgf("%s", line)
}

func printOtherWeeks(day int, start time.Time, schedule pagerduty.Overrides) {
	e := endOfMonth(start)
	idx := 0
	line := ""
	for day <= e.Day() {
		line += printDay(day, idx, start)
		line += printName(day, schedule)
		idx++

		if idx >= len(days) {
			log.Info().Msgf("%s", line)
			line = ""
			idx = 0
		}

		day++
	}
}

func printName(idx int, schedule pagerduty.Overrides) string {
	firstName := strings.Split(schedule.Overrides[idx-1].User.Name, " ")[0]
	if len(firstName) > nameLen {
		firstName = firstName[:nameLen]
	} else {
		firstName += strings.Repeat(" ", nameLen-len(firstName))
	}

	return firstName
}

func printDay(day, idx int, start time.Time) string {
	workdayColor := color.New(color.FgWhite).Add(color.Bold)
	holidayColor := color.New(color.FgHiCyan).Add(color.Bold)
	currentDayColor := color.New(color.FgHiRed).Add(color.Bold)

	if day > 9 { //nolint:gomnd // obvious value
		if day == start.Day() {
			return currentDayColor.Sprintf(" %d%s", day, daySeparator)
		} else if idx == 5 || idx == 6 {
			return holidayColor.Sprintf(" %d%s", day, daySeparator)
		} else {
			return workdayColor.Sprintf(" %d%s", day, daySeparator)
		}
	} else {
		if day == start.Day() {
			return currentDayColor.Sprintf("  %d%s", day, daySeparator)
		} else if idx == 5 || idx == 6 {
			return holidayColor.Sprintf("  %d%s", day, daySeparator)
		} else {
			return workdayColor.Sprintf("  %d%s", day, daySeparator)
		}
	}
}

func beginningOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 0, -date.Day()+1)
}

func endOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 1, -date.Day())
}
