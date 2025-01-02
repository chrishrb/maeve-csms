.PHONY: up
up: register-cs setup-rfid setup-contract
	OCPP_VERSION=ocpp docker-compose --profile everest up -d

.PHONY: down
down:
	docker-compose down

.PHONY: test
test: # password 123456
	websocat --client-pkcs12-der ./config/certificates/cs001.pem --client-pkcs12-passwd 123456 wss://localhost:443

.PHONY: setup-rfid
setup-rfid:
	curl -i http://localhost:9410/api/v0/token -H 'content-type: application/json' -d '{"country_code": "GB","party_id": "TWK","type": "RFID","uid": "DEADBEEF","contract_id": "GBTWK012345678V","issuer": "Thoughtworks","valid": true,"cache_mode": "ALWAYS"}'

.PHONY: setup-contract
setup-contract:
	curl -i http://localhost:9410/api/v0/token -H 'content-type: application/json' -d '{"country_code": "GB","party_id": "TWK","type": "RFID","uid": "EMP77TWTW99999","contract_id": "GBTWK012345678V","issuer": "Thoughtworks","valid": true,"cache_mode": "ALWAYS"}'

.PHONY: register-cs
register-cs:
	bash everest/scripts/setup-env.sh
