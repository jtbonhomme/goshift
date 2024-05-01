# goshift

This repository allows on-call teams to build multi layered monthly schedules based on engineers availabilities, and various criterias.

## Context and motivation

Building on-call schedules may be tedious when you need to deal with several constraints like seniority, availabilities, week-end management, and so on.
This repository implements simples rules to automatically build primary and secondary on-call planning that comply with those rules with optimized fairness (eg try to dispatch evenly week days and week-end on-call shifts)

The program is highly coupled to PagerDuty data structures and it builds schedules as PagerDuty API compatible JSON objects.

## Features

* On-Call Engineers fills their monthly availabilities in planning service (framadate.org is currently supported),
* Availabilities are downloaded as a CSV file,
* The program is launched with various input files (framadate CSV avaliabilities, list of junior engineers that should not be in secondary schedules, list of resgistrated users, list of last engineers that were on-call last day of previous month),
* 2 JSON object files are proposed that can be curl-ed directly to PagerDuty to create month overrides.

## Internal algorithm

The program starts by ingesting various input files and build a list of engineers(users) with their availabilities.
Then, the `solver` is run to produce primary and secondary override schedules.

The two main criterai used to select user during the override schedule building are:
* non repetitive selection criteria
* even distribution of on-call shifts (aka fairness criteria)

The solver algorithm could be described with this pseudo-code:
```
intialize primary and seconday overrides
initialize last assigned users (1)
for day in month do
      get list of users, sorted by their reverse number of remaining availabilities (2)
      select next available user for primary override matching both non repetitive selection criteria and even distribution of on-call shifts (3 & 4)
      select next available user for secondary override matching both non repetitive selection criteria and even distribution of on-call shifts (3 & 4)
      if selection for primary and secondary is empty (5)
            then select fist available user matching only non repetitive selection criteria
end for
```

(1) **last assigned users** are intiailized with either the assigned engineers of the last day of the previous month before the loop starts, or the assigned engineers of the previous day of the current month inside the loop.
(2) **less available user sorting** the goal of this sorting is to foster user selection as soon as possible when they have less availabilities in the next days.
(3) **non repetitive selection criteria** to avoid the same person being on-call two consecutive days. Week-ends are by limited to Saturday/Sunday (see limitations).
(4) **even distribution of on-call shifts (aka fairness criteria)**  we try to distribute number of on-call shifts every month over engineers regardless of their availabilities. Of course, it is only an optimization attempt, even distribution of week days and week-end in not guaranted.
(5) **empty selection** happens when no user mating both  **non repetitive selection criteria** and **even distribution of on-call shifts (aka fairness criteria)** have been found.

**Note**: junior (newbies) users can not be selected for secondary schedules

## Usage

```sh
Usage of goshift:
  -csv string
        [mandatory] framadate csv file path
  -debug
        sets log level to debug
  -last value
        [optional] last users emails of previous schedule. Emails must match users json file.
  -newbies string
        [optional] newbies json file path")
  -users string
        [mandatory] users json file path")
```

## Get Started

1. Ask oncall people to fill a framadate calendar for the month (example: https://framadate.org/2M5jvRhBTCshJycz)
2. Download schedule as CSV from Framadate
3. Download pagerduty schedule users from API: 
```sh
curl -s -o ~/Documents/pagerduty-users.json --request GET \
  --url https://api.pagerduty.com/schedules/<SCHEDULE-ID>/users \
  --header 'Accept: application/json' \
  --header 'Authorization: Token token=<API-KEY>' \
  --header 'Content-Type: application/json'
```
4. Fill the newbies JSON file
5. Run `goshift`:

```sh
go run cmd/goshift/main.go  -users ~/Documents/pagerduty-users.json -csv ~/Downloads/On-CallMay2024.csv -last user1@email.com -last user2@email.com -debug      
```
This will create two files `primary.json` and `secondary.json`

6. A report is output to let the operator to quickly control the schedules and the distribution of week days and weekend shifts.

```
11:32AM INF Primary on-call shift
11:32AM INF
11:32AM INF May 2024
11:32AM INF
11:32AM INF Mon           Tue           Wed           Thu           Fri           Sat           Sun
11:32AM INF                               1 User9       2 User2       3 User8       4 User2       5 User2       
11:32AM INF   6 User6       7 User10      8 User4       9 User10     10 User4      11 User10     12 User10   
11:32AM INF  13 User5      14 User6      15 User8      16 User6      17 User8      18 User1      19 User1     
11:32AM INF  20 User4      21 User8      22 User4      23 User8      24 User6      25 User4      26 User4     
11:32AM INF
11:32AM INF
11:32AM INF Secondary on-call shift
11:32AM INF
11:32AM INF May 2024
11:32AM INF
11:32AM INF Mon           Tue           Wed           Thu           Fri           Sat           Sun
11:32AM INF                               1 User7       2 User3       3 User1       4 User7       5 User7     
11:32AM INF   6 User1       7 User3       8 User1       9 User3      10 User1      11 User3      12 User3     
11:32AM INF  13 User9      14 User7      15 User5      16 User9      17 User5      18 User7      19 User7     
11:32AM INF  20 User5      21 User9      22 User5      23 User9      24 User3      25 User5      26 User5     
11:32AM INF
11:32AM INF
11:32AM INF
11:32AM INF +--------------------------------------------------------------+----+----+----+----+
11:32AM INF | Email                                                        |  S |  W |   u | v |
11:32AM INF +--------------------------------------------------------------+----+----+----+----+
11:32AM INF | user1@email.com                                              |  7 |  1 |  1 |  0 |
11:32AM INF | user2@email.com                                              |  6 |  1 | 16 |  3 |
11:32AM INF | user3@email.com                                              |  7 |  1 |  5 |  0 |
11:32AM INF | user4@email.com                                              |  6 |  1 |  0 |  0 |
11:32AM INF | user5@email.com                                              |  7 |  1 | 10 |  2 |
11:32AM INF | user6@email.com                                              |  6 |  0 | 13 |  3 |
11:32AM INF | user7@email.com                                              |  7 |  2 |  7 |  1 |
11:32AM INF | user8@email.com                                              |  5 |  0 |  9 |  4 |
11:32AM INF | user9@email.com                                              |  7 |  0 |  7 |  4 |
11:32AM INF | user10@email.com                                             |  4 |  1 | 10 |  1 |
11:32AM INF +--------------------------------------------------------------+----+----+----+----+
```

7. Post the override schedules to pagerduty:

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

## Limitations

* Week-end definition is limited to Saturday/Sunday

## ToDo

* [ ] Use Teams / Members PD api
```sh
curl -s -o ~/Documents/pagerduty-members.json --request GET  --url https://api.pagerduty.com/teams/<TEAM-ID>/members  --header 'Accept: application/json'  --header 'Authorization: Token token=<API-KEY>'  --header 'Content-Type: application/json'
```
* [ ] Manage geos of engineers to use local week-end definition