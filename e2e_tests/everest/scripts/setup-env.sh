#!/usr/bin/env bash

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

LOCATION_ID=$(curl -X 'POST' \
  'http://localhost:9410/api/v0/location' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "address": "Examplestreet 123",
    "city": "Examplecity",
    "country": "Germany",
    "country_code": "DE",
    "coordinates": {
      "latitude": "51.047599",
      "longitude": "3.729944"
    },
    "party_id": "1234"
  }' | jq -r '.id')
echo "$LOCATION_ID"

CS_ID=$(curl -X 'POST' \
  'http://localhost:9410/api/v0/cs' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d "{\"security_profile\":2,\"location_id\":\"${LOCATION_ID}\"}" \
  | jq -r '.id')

sed "s/<CS_ID>/${CS_ID}/" "$SCRIPT_DIR/../config/everest/ocpp/OCPP/config.template.json" > "$SCRIPT_DIR/../config/everest/ocpp/OCPP/config.json"
sed "s/<CS_ID>/${CS_ID}/" "$SCRIPT_DIR/../config/everest/ocpp/OCPP201/component_config/standardized/InternalCtrlr.template.json" > "$SCRIPT_DIR/../config/everest/ocpp/OCPP201/component_config/standardized/InternalCtrlr.json"
sed "s/<CS_ID>/${CS_ID}/" "$SCRIPT_DIR/../config/everest/ocpp/OCPP201/component_config/standardized/SecurityCtrlr.template.json" > "$SCRIPT_DIR/../config/everest/ocpp/OCPP201/component_config/standardized/SecurityCtrlr.json"

echo "cs with id ${CS_ID} created"
