package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/jtbonhomme/goshift/internal/pagerduty"
	"github.com/jtbonhomme/goshift/internal/schedule"
	"github.com/jtbonhomme/goshift/internal/solver"
	"github.com/jtbonhomme/goshift/internal/utils"
)

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
	var users pagerduty.Users

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(usersValue, &users)

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

	input := utils.ParseFramadateCSV(data)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(input.Users), func(i, j int) { input.Users[i], input.Users[j] = input.Users[j], input.Users[i] })

	primary, secondary, pstats, sstats, err := solver.Run(input, users)
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

	schedule.DisplayCalendar("Primary on-call shift", primary)

	s, err := json.MarshalIndent(secondary, "", "  ")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("secondary.json", s, 0644)
	if err != nil {
		panic(err)
	}

	schedule.DisplayCalendar("Secondary on-call shift", secondary)
}
