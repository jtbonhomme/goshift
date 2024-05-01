package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"io"
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

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "string slice representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var err error
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("goshift")
	var csvPath, usersPath, newbiesPath string
	var debug bool
	var lastUsers arrayFlags

	flag.BoolVar(&debug, "debug", false, "sets log level to debug")
	flag.StringVar(&csvPath, "csv", "", "[mandatory] framadate csv file path")
	flag.StringVar(&usersPath, "users", os.Getenv("HOME")+"/Documents/pagerduty-users.json", "[optional] users json file path")
	flag.StringVar(&newbiesPath, "newbies", os.Getenv("HOME")+"/Documents/pagerduty-newbies.json", "[optional] newbies json file path")
	flag.Var(&lastUsers, "last", "[optional] last users emails of previous schedule")

	flag.Parse()

	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if csvPath == "" {
		panic(errors.New("framadate csv file is missing"))
	}

	usersJson, err := os.Open(usersPath)
	if err != nil {
		panic(errors.New("unable to open users file " + usersPath + " : " + err.Error()))
	}
	log.Info().Msg("Successfully opened users.json")
	defer usersJson.Close()
	usersValue, err := io.ReadAll(usersJson)
	if err != nil {
		panic(errors.New("unable to read json file : " + err.Error()))
	}

	var users pagerduty.Users
	err = json.Unmarshal(usersValue, &users)
	if err != nil {
		panic(errors.New("unable to unmarshall users JSON value: " + err.Error()))
	}

	newbiesJson, err := os.Open(newbiesPath)
	if err != nil {
		panic(errors.New("unable to open newbies file " + newbiesPath + " : " + err.Error()))
	}

	log.Info().Msg("Successfully opened newbies.json")
	defer newbiesJson.Close()
	newbiesValue, err := io.ReadAll(newbiesJson)
	if err != nil {
		panic(errors.New("unable to read json file : " + err.Error()))
	}

	var newbies []string
	err = json.Unmarshal(newbiesValue, &newbies)
	if err != nil {
		panic(errors.New("unable to unmarshall newbies JSON value: " + err.Error()))
	}

	f, err := os.Open(csvPath)
	if err != nil {
		panic(errors.New("unable to open csv file " + csvPath + " : " + err.Error()))
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		panic(errors.New("unable to read csv file : " + err.Error()))
	}

	input := utils.ParseFramadateCSV(data)

	unavailablitiesStats := input.UnavailablitiesStats()

	sv := solver.New(input, users, newbies, []string(lastUsers))
	primary, secondary, err := sv.Run()
	if err != nil {
		panic(err)
	}

	log.Info().Msg("")

	p, err := json.MarshalIndent(primary, "", "  ")
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("primary.json", p, 0644)
	if err != nil {
		panic(err)
	}

	schedule.DisplayCalendar("Primary on-call shift", primary)

	s, err := json.MarshalIndent(secondary, "", "  ")
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("secondary.json", s, 0644)
	if err != nil {
		panic(err)
	}

	schedule.DisplayCalendar("Secondary on-call shift", secondary)

	log.Info().Msg("")

	h := color.New(color.FgHiBlue).Add(color.Bold)
	log.Info().Msgf("+%s+----+----+----+----+", strings.Repeat("-", 62))
	log.Info().Msgf("| %s                                                        |  %s |  %s |   %s | %s |", h.Sprint("Email"), h.Sprint("S"), h.Sprint("W"), h.Sprint("u"), h.Sprint("v"))
	log.Info().Msgf("+%s+----+----+----+----+", strings.Repeat("-", 62))

	for _, user := range input.Users {
		log.Info().Msgf("| %s %s| %2d | %2d | %2d | %2d |", user.Email, strings.Repeat(" ", 60-len(user.Email)), sv.Stats[user.Email], sv.WeekendStats[user.Email], unavailablitiesStats.Weekdays[user.Email], unavailablitiesStats.Weekends[user.Email])
	}
	log.Info().Msgf("+%s+----+----+----+----+", strings.Repeat("-", 62))
	log.Info().Msg("")
}
