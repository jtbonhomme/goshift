package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/jtbonhomme/goshift/internal/pagerduty"
	"github.com/jtbonhomme/goshift/internal/schedule"
	"github.com/jtbonhomme/goshift/internal/solver"
	"github.com/jtbonhomme/goshift/internal/utils"
)

func main() {
	var err error
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("goshift")
	var csvPath string
	var debug bool
	flag.BoolVar(&debug, "debug", false, "sets log level to debug")
	flag.StringVar(&csvPath, "csv", "", "[mandatory] framadate csv file path")
	flag.Parse()

	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if csvPath == "" {
		panic(errors.New("framadate csv file is missing"))
	}

	usersJson, err := os.Open("/Users/jean-thierry.bonhomme/Documents/Contentsquare/pagerduty-users.json")
	if err != nil {
	}
	log.Info().Msg("Successfully opened users.json")
	defer usersJson.Close()
	usersValue, _ := ioutil.ReadAll(usersJson)

	var users pagerduty.Users
	json.Unmarshal(usersValue, &users)

	newbiesJson, err := os.Open("/Users/jean-thierry.bonhomme/Documents/Contentsquare/pagerduty-newbies.json")
	if err != nil {
	}
	log.Info().Msg("Successfully opened newbies.json")
	defer newbiesJson.Close()
	newbiesValue, _ := ioutil.ReadAll(newbiesJson)

	var newbies []string
	json.Unmarshal(newbiesValue, &newbies)

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

	sv := solver.New(input, users, newbies)
	primary, secondary, err := sv.Run()
	if err != nil {
		panic(err)
	}

	log.Info().Msg("")

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

	log.Info().Msg("")

	h := color.New(color.FgHiBlue).Add(color.Bold)
	log.Info().Msgf("+%s+----+----+", strings.Repeat("-", 62))
	log.Info().Msgf("| %s                                                        |  %s |  %s |", h.Sprint("Email"), h.Sprint("S"), h.Sprint("W"))
	log.Info().Msgf("+%s+----+----+", strings.Repeat("-", 62))

	for _, user := range input.Users {
		log.Info().Msgf("| %s %s| %2d | %2d |", user.Email, strings.Repeat(" ", 60-len(user.Email)), sv.Stats[user.Email], sv.WeekendStats[user.Email])
	}
	log.Info().Msgf("+%s+----+----+", strings.Repeat("-", 62))
	log.Info().Msg("")
}
