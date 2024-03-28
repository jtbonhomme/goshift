package pagerduty

import (
	"fmt"
	"slices"
	"time"
)

// User have a name, id, type, unavailable dates, and preferences.
type User struct {
	Name        string      `json:"name,omitempty"`
	Email       string      `json:"email,omitempty"`
	ID          string      `json:"id,omitempty"`
	Type        string      `json:"type,omitempty"`
	Unavailable []time.Time `json:"unavailable,omitempty"`
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

func (ui *UserIterator) Len(exclude []string) int {
	if exclude == nil {
		return len(ui.Users)
	}

	return len(ui.Users) - len(exclude)
}

func (ui *UserIterator) NextWithExclude(exclude []string) (User, int, bool) {
	if exclude == nil {
		u, n := ui.Next()
		return u, n, true
	}

	var found bool
	var k int

	for i := 0; i < len(ui.Users); i++ {
		k = ui.iterator % len(ui.Users)
		ui.iterator++
		if !slices.Contains(exclude, ui.Users[k].Email) {
			found = true
			break
		}
	}

	return ui.Users[k], k, found
}

func (ui *UserIterator) Next() (User, int) {
	k := ui.iterator % len(ui.Users)
	ui.iterator++

	return ui.Users[k], k
}

func (users Users) RetrieveAssignedUser(user User) (AssignedUser, error) {
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

func (users Users) RetrieveAssignedUserByEmail(email string) (AssignedUser, error) {
	for _, u := range users.Users {
		if u.Email == email {
			return AssignedUser{
				Name:  u.Name,
				Email: u.Email,
				ID:    u.ID,
				Type:  u.Type,
			}, nil
		}
	}

	return AssignedUser{}, fmt.Errorf("unknown user %s", email)
}
