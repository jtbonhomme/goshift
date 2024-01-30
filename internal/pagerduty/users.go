package pagerduty

import (
	"fmt"
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

func (ui *UserIterator) Next() (User, int) {
	k := ui.iterator % len(ui.Users)
	ui.iterator++
	return ui.Users[k], k
}

func RetrieveUser(user User, users Users) (AssignedUser, error) {
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
