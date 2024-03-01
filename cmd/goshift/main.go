package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
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

var newbies []string = []string{
	"valerio.figliuolo@contentsquare.com",
	"ahmed.khaled@contentsquare.com",
	"houssem.touansi@contentsquare.com",
	"kevin.albes@contentsquare.com",
	"yunbo.wang@contentsquare.com",
	"wael.tekaya@contentsquare.com",
}

func main() {
	var err error

	fmt.Println("goshift")
	var csvPath string
	var noShuffle bool
	flag.StringVar(&csvPath, "csv", "", "[mandatory] framadate csv file path")
	flag.BoolVar(&noShuffle, "n", false, "[optional] do not shuffle users before creating shifts")
	flag.Parse()

	if csvPath == "" {
		panic(errors.New("framadate csv file is missing"))
	}

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

	f, err := os.Open(csvPath)
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
	if !noShuffle {
		rand.Shuffle(len(input.Users), func(i, j int) { input.Users[i], input.Users[j] = input.Users[j], input.Users[i] })
	}

	sv := solver.New(input, users, newbies)
	primary, secondary, err := sv.Run()
	if err != nil {
		panic(err)
	}

	fmt.Println("")

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

	fmt.Println("")

	fmt.Printf("+%s+----+----+\n", strings.Repeat("-", 62))
	fmt.Println("| Email                                                        |  S |  W |")
	fmt.Printf("+%s+----+----+\n", strings.Repeat("-", 62))

	for _, user := range input.Users {
		fmt.Printf("| %s %s| %2d | %2d |\n", user.Email, strings.Repeat(" ", 60-len(user.Email)), sv.Stats[user.Email], sv.WeekendStats[user.Email])
	}
	fmt.Printf("+%s+----+----+\n", strings.Repeat("-", 62))
	fmt.Println("")
}
