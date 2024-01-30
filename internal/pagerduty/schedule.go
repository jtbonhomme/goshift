package pagerduty

import (
	"time"
)

// Input for the pager duty scheduling problem. We have
// pager duty users that need to be assigned to days between the schedule start
// date and the schedule end date.
type Input struct {
	ScheduleStart time.Time `json:"schedule_start"`
	ScheduleEnd   time.Time `json:"schedule_end"`
	Users         []User    `json:"users"`
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
