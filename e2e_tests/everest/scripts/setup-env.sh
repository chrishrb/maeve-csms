#!/usr/bin/env bash

curl -X 'POST' \
  'http://localhost:9410/api/v0/location' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "id": "loc001"
    "address": "Examplestreet 123",
    "city": "Examplecity",
    "country": "Germany",
    "country_code": "DE",
    "coordinates": {
      "latitude": "51.047599",
      "longitude": "3.729944"
    },
    "party_id": "1234"
  }'

curl -X 'POST' \
  'http://localhost:9410/api/v0/cs' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{"id": "cp001","security_profile":2,"location_id":"loc001"}'

echo "cs with id cp001 created"
