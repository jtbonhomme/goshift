package solver

import (
	"testing"
	"time"

	"github.com/jtbonhomme/goshift/internal/pagerduty"
	"github.com/jtbonhomme/goshift/internal/utils"
)

/*
March 2024
Mon           Tue           Wed           Thu           Fri           Sat           Sun
 11           12            13            14            15            16            17
              start         +1            +2            +3            +4            +5
*/

func TestNextAvailableUser(t *testing.T) {
	var err error
	var location *time.Location
	location, err = time.LoadLocation("Europe/Paris")
	if err != nil {
		panic(err)
	}

	start, err := time.Parse("02/01/2006", "03/12/2024") // 12th Mar. 2024 is a Tuesday
	if err != nil {
		panic(err)
	}

	tstart := time.Date(start.Year(), start.Month(), start.Day(), 9, 0, 0, 0, location)

	users := []pagerduty.User{
		{Name: "user0", Email: "user0@test", Unavailable: []time.Time{tstart.Add(utils.OneDay), tstart.Add(2 * utils.OneDay), tstart.Add(3 * utils.OneDay)}},
		{Name: "user1", Email: "user1@test", Unavailable: []time.Time{tstart.Add(2 * utils.OneDay), tstart.Add(3 * utils.OneDay)}},
		{Name: "user2", Email: "user2@test", Unavailable: []time.Time{tstart.Add(3 * utils.OneDay)}},
		{Name: "user3", Email: "user3@test", Unavailable: []time.Time{tstart.Add(3 * utils.OneDay)}},
	}

	input := pagerduty.Input{
		ScheduleStart: tstart,
		ScheduleEnd:   tstart.Add(3 * utils.OneDay),
		Users:         users,
	}

	s := New(input, pagerduty.Users{Users: users})
	d := input.ScheduleStart
	var u pagerduty.User
	var n int
	var ok bool

	u, n, ok = s.nextAvailableUser(d)
	if !ok {
		t.Errorf("ok expected to be true but got %v", ok)
	}
	if n != 0 {
		t.Errorf("n expected to be 0 but got %d", n)
	}
	if u.Name != "user0" {
		t.Errorf("user name expected to be user0 but got %s", u.Name)
	}

	u, n, ok = s.nextAvailableUser(d.Add(utils.OneDay))
	if !ok {
		t.Errorf("ok expected to be true but got %v", ok)
	}
	if n != 1 {
		t.Errorf("n expected to be 1 but got %d", n)
	}
	if u.Name != "user1" {
		t.Errorf("user name expected to be user1 but got %s", u.Name)
	}

	u, n, ok = s.nextAvailableUser(d.Add(2 * utils.OneDay))
	if !ok {
		t.Errorf("ok expected to be true but got %v", ok)
	}
	if n != 2 {
		t.Errorf("n expected to be 2 but got %d", n)
	}
	if u.Name != "user2" {
		t.Errorf("user name expected to be user2 but got %s", u.Name)
	}

	u, n, ok = s.nextAvailableUser(d.Add(3 * utils.OneDay))
	if ok {
		t.Errorf("ok expected to be false but got %v", ok)
	}
}

/*
March 2024
Mon           Tue           Wed           Thu           Fri           Sat           Sun
 11           12            13            14            15            16            17
              start         +1            +2            +3            +4            +5
*/

func TestNextAvailableUserWeekend(t *testing.T) {
	var err error
	var location *time.Location
	location, err = time.LoadLocation("Europe/Paris")
	if err != nil {
		panic(err)
	}

	start, err := time.Parse("02/01/2006", "03/12/2024") // 12th Mar. 2024 is a Tuesday
	if err != nil {
		panic(err)
	}

	tstart := time.Date(start.Year(), start.Month(), start.Day(), 9, 0, 0, 0, location)

	users := []pagerduty.User{
		{Name: "user0", Email: "user0@test", Unavailable: []time.Time{tstart.Add(utils.OneDay), tstart.Add(2 * utils.OneDay), tstart.Add(3 * utils.OneDay), tstart.Add(4 * utils.OneDay)}},
		{Name: "user1", Email: "user1@test", Unavailable: []time.Time{tstart.Add(2 * utils.OneDay), tstart.Add(3 * utils.OneDay), tstart.Add(4 * utils.OneDay)}},
		{Name: "user4", Email: "user4@test", Unavailable: []time.Time{tstart.Add(3 * utils.OneDay), tstart.Add(4 * utils.OneDay)}}, // unavailabe Saturday
		{Name: "user5", Email: "user5@test", Unavailable: []time.Time{tstart.Add(3 * utils.OneDay), tstart.Add(5 * utils.OneDay)}}, // unavailabe Sunday
	}

	input := pagerduty.Input{
		ScheduleStart: tstart,
		ScheduleEnd:   tstart.Add(3 * utils.OneDay),
		Users:         users,
	}

	s := New(input, pagerduty.Users{Users: users})
	d := input.ScheduleStart
	var u pagerduty.User
	var n int
	var ok bool

	u, n, ok = s.nextAvailableUser(d)
	if !ok {
		t.Errorf("ok expected to be true but got %v", ok)
	}
	if n != 0 {
		t.Errorf("n expected to be 0 but got %d", n)
	}
	if u.Name != "user0" {
		t.Errorf("user name expected to be user0 but got %s", u.Name)
	}

	s = New(input, pagerduty.Users{Users: users})
	u, n, ok = s.nextAvailableUser(d.Add(4 * utils.OneDay))
	if ok {
		t.Errorf("ok expected to be false but got %v", ok)
	}
}
