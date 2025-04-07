# execute these tasks when `make` with no target is invoked
default: unit-test reset ready e2e-test logs

# import environment file for setting or overriding
# configuration used by this Makefile
include .env

# source all variables in environment file
# This only runs in the make command shell
# so won't affect your login shell
export $(shell sed 's/=.*//' .env)

# Make functions that can be called inside of targets to check
# that required environment variables are set, e.g.
check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))

.PHONY: lint
# format and verify service source code and dependency tree
lint:
	cd ${NEXUS_API_DIRECTORY} && \
	go mod tidy && \
	go fmt ./ && \
	go vet ./

.PHONY: build
# build a development version docker image of the service
build: lint
	docker compose build

# .PHONY: publish
# # build a production version docker image of the service
# publish: lint
# 	docker build ./ -f production.Dockerfile -t ${IMAGE_NAME}:${PRODUCTION_IMAGE_TAG}

.PHONY: unit-test
# run all unit tests
unit-test:
	cd ${NEXUS_API_DIRECTORY} && \
	go test -count=1 -v -cover -coverprofile cover.out --race ./... -run "^TestUnitTest*"

.PHONY: e2e-test
# run tests that execute against a local or remote instance of the API
e2e-test:
	cd ${NEXUS_API_DIRECTORY} && \
	go test -p=1 -count=1 -v -cover -coverprofile cover.out --race ./... -run "^TestE2ETest*"

.PHONY: it
# run any test matching the provided pattern, can pass a regex or a string
# of the exact test to run
it : lint
	cd ${NEXUS_API_DIRECTORY} && \
	go test -count=1 -v -cover -coverprofile cover.out --race ./... -run=".*${p}.*"

.PHONY: show-coverage
# convert test coverage report to html & open in browser
show-coverage:
	cd ${NEXUS_API_DIRECTORY} && \
	go tool cover -html cover.out -o cover.html && open cover.html

.PHONY: test
# run all tests
test:
	cd ${NEXUS_API_DIRECTORY} && \
	go test -count=1 -v -cover -coverprofile cover.out --race ./...

.PHONY: up
# start dockerized versions of the service and it's dependencies
up:
	docker compose up -d

.PHONY: down
# stop the service and it's dependencies
down:
	docker compose down

.PHONY: restart
# restart just the service (useful for picking up new environment variable values)
restart:
	docker compose up -d nexus-api --force-recreate

.PHONY: reset
# wipe state and restart the service and all it's dependencies
reset: lint
	docker compose up -d --build --remove-orphans --renew-anon-volumes --force-recreate

.PHONY: refresh
# rebuild from latest local sources and restart just the service containers
# (preserving any volume state such as database tables & rows)
refresh: lint
	docker compose up -d nexus-api --build --force-recreate

.PHONY: logs
# follow the logs from all the dockerized services
# make logs
# or one
# make logs S=nexus-api
logs:
	docker compose logs -f ${S}

# poll api health check endpoint until it doesn't error
.PHONY: ready
ready:
	./bin/wait_for_api_ready.sh

.PHONY: debug-nexus-api
# attach the dlv debugger to the running service and connect to the dlv debugger
debug-nexus-api:
	docker compose exec -d nexus-api dlv attach 1 --listen=:${NEXUS_API_CONTAINER_DEBUG_PORT} --headless --api-version=2 --log && \
	dlv connect :${NEXUS_API_HOST_DEBUG_PORT}

.PHONY: debug-database
# open a connection to the postgres database for debugging it's state
# https://www.postgresql.org/docs/current/app-psql.html
debug-database:
	docker compose exec nexus-db psql -U ${POSTGRES_USER} -d ${POSTGRES_DB}

.PHONY: seed-database
# add seed state to database
seed-database:
	PGPASSWORD=${POSTGRES_PASSWORD} psql -d ${POSTGRES_DB} -U ${POSTGRES_USER} -h localhost -f nexus-api/seed/local_login_authentication_seed.sql

	PGPASSWORD=${POSTGRES_PASSWORD} psql -d ${POSTGRES_DB} -U ${POSTGRES_USER} -h localhost -f nexus-api/seed/local_panel_yield_data_seed.sql

	PGPASSWORD=${POSTGRES_PASSWORD} psql -d ${POSTGRES_DB} -U ${POSTGRES_USER} -h localhost -f nexus-api/seed/local_panel_consumption_data_seed.sql

	PGPASSWORD=${POSTGRES_PASSWORD} psql -d ${POSTGRES_DB} -U ${POSTGRES_USER} -h localhost -f nexus-api/seed/local_solar_consumption_data_seed.sql
	
	PGPASSWORD=${POSTGRES_PASSWORD} psql -d ${POSTGRES_DB} -U ${POSTGRES_USER} -h localhost -f nexus-api/seed/local_sensor_moisture_data_seed.sql

	PGPASSWORD=${POSTGRES_PASSWORD} psql -d ${POSTGRES_DB} -U ${POSTGRES_USER} -h localhost -f nexus-api/seed/local_sensor_temperature_data_seed.sql

	PGPASSWORD=${POSTGRES_PASSWORD} psql -d ${POSTGRES_DB} -U ${POSTGRES_USER} -h localhost -f nexus-api/seed/local_sensor_data_seed.sql
	

