package pagerduty

import (
	"fmt"
	"testing"
)

func TestNext(t *testing.T) {
	users := []User{
		{Name: "user0"},
		{Name: "user1"},
		{Name: "user2"},
	}
	ui := NewIterator(users)
	for i := 0; i <= 15; i++ {
		u, n := ui.Next()
		fmt.Println(i, u.Name, n)
	}
}
