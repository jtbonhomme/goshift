

curl -s -o ~/Documents/Contentsquare/pagerduty-users.json --request GET \
  --url https://api.pagerduty.com/schedules/P0117U4/users \
  --header 'Accept: application/json' \
  --header 'Authorization: Token token=<API-KEY>' \
  --header 'Content-Type: application/json'


  curl --request POST --url https://api.pagerduty.com/schedules/P0117U4/overrides \
  --header 'Accept: application/json' \
  --header 'Authorization: Token token=<API-KEY>' \
  --header 'Content-Type: application/json' \
  --data @primary.json


  curl --request POST --url https://api.pagerduty.com/schedules/P9VKKU3/overrides \
  --header 'Accept: application/json' \
  --header 'Authorization: Token token=<API-KEY>' \
  --header 'Content-Type: application/json' \
  --data @secondary.json