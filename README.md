# goshift

## Usage

```sh
Usage of goshift:
  -csv string
        [mandatory] framadate csv file path
  -debug
        sets log level to debug
  -last value
        [optional] last users emails of previous schedule
  -newbies string
        [optional] newbies json file path")
  -users string
        [optional] users json file path")
```

## Get Started

1. Ask oncall people to fill a framadate calendar for the month (example: https://framadate.org/2M5jvRhBTCshJycz)
2. Download schedule as CSV from Framadate
3. Download pagerduty schedule users from API: 
```sh
curl -s -o ~/Documents/Contentsquare/pagerduty-users.json --request GET \
  --url https://api.pagerduty.com/schedules/<SCHEDULE-ID>/users \
  --header 'Accept: application/json' \
  --header 'Authorization: Token token=<API-KEY>' \
  --header 'Content-Type: application/json'
```
4. Run `goshift`:

Update newbie list if needed in `internal/solver/solver.go`

Then run:
```sh
go run cmd/goshift/main.go
```

This will create two files `primary.json` and `secondary.json`

5. Post the override schedules to pagerduty:

```sh
  curl --request POST --url https://api.pagerduty.com/schedules/<PRIMARY-SCHEDULE-ID>/overrides \
  --header 'Accept: application/json' \
  --header 'Authorization: Token token=<API-KEY>' \
  --header 'Content-Type: application/json' \
  --data @primary.json

  curl --request POST --url https://api.pagerduty.com/schedules/<SECONDARY-SCHEDULE-ID>/overrides \
  --header 'Accept: application/json' \
  --header 'Authorization: Token token=<API-KEY>' \
  --header 'Content-Type: application/json' \
  --data @secondary.json
```

## ToDO

Use Teams / Members api:
```sh
curl -s -o ~/Documents/Contentsquare/pagerduty-members.json --request GET  --url https://api.pagerduty.com/teams/<TEAM-ID>/members  --header 'Accept: application/json'  --header 'Authorization: Token token=<API-KEY>'  --header 'Content-Type: application/json'
``` 