# ==================================================================================== #
# VARIABLES
# ==================================================================================== #
# Include variables from the local .env file
include ./app.env

VERSION := 1.22
UID := $(shell id -u)
GID := $(shell id -g)

APP_MODULE := github.com/Housiadas/simple-banking-system
INPUT ?= $(shell bash -c 'read -p "Insert name: " name; echo $$name')
CURRENT_TIME := $(shell date --iso-8601=seconds)
GIT_VERSION := $(shell git describe --always --dirty --tags --long)
LINKER_FLAGS := "-s -X main.buildTime=${CURRENT_TIME} -X main.version=${GIT_VERSION}"

DOCKER_COMPOSE_LOCAL := docker-compose -f ./docker-compose.yml
MIGRATE := $(DOCKER_COMPOSE_LOCAL) run --rm migrate
SQLC := $(DOCKER_COMPOSE_LOCAL) run --rm sqlc

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## docker/build: Build the application
.PHONY: docker/build
docker/build:
	docker build --target application -t banking-api:local \
		--build-arg VERSION=$(VERSION) \
 		-f .docker/app/Dockerfile .

## docker/up: Start all the containers for the application
.PHONY: docker/up
docker/up:
	$(DOCKER_COMPOSE_LOCAL) up -d db redis

## docker/stop: stop all containers
.PHONY: docker/stop
docker/stop:
	$(DOCKER_COMPOSE_LOCAL) stop

## docker/down: stop and remove all containers
.PHONY: docker/down
docker/down:
	$(DOCKER_COMPOSE_LOCAL) down --remove-orphans

## docker/clean: docker clean all
.PHONY: docker/clean
docker/clean:
	docker system prune && \
    docker image prune && \
    docker volume prune

## go/mock/store: Go mock Store interface
.PHONY: go/mock/store
go/mock/store:
	mockgen -package mockdb -destination business/db/mock/store.go $(APP_MODULE)/business/db Store

## go/run: Run main.go locally
.PHONY: go/run
go/run:
	go run app/main.go

# ==================================================================================== #
# DATABASE
# ==================================================================================== #

## db/migrations/create name=$1: create new migration files
.PHONY: db/migrate/create
db/migrate/create:
	$(MIGRATE) create -seq -ext=.sql -dir=./database/migrations $(INPUT)

## db/migrations/up: apply all up database migrations
.PHONY: db/migrate/up
db/migrate/up:
	$(MIGRATE) -path=./database/migrations -database=${MIGRATION_DB_DSN} up

## db/migrations/down: apply all down database migrations (DROP Database)
.PHONY: db/migrate/down
db/migrate/down:
	$(MIGRATE) -path=./database/migrations -database=${MIGRATION_DB_DSN} down

## db/migrations/local/up: apply all up database migrations local command
.PHONY: db/migrate/local/up
db/migrate/local/up:
	go -path=./database/migrations -database=${MIGRATION_DB_DSN} up

## db/sqlc/init: Create an empty sqlc.yaml settings file
.PHONY: db/sqlc/init
db/sqlc/init:
	$(SQLC) init

## db/sqlc/init: Create an empty sqlc.yaml settings file
.PHONY: db/sqlc/generate
db/sqlc/generate:
	$(SQLC) generate

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

# update: update dependencies
.PHONY: update
update:
	go get -u ./...
	go mod verify

# vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor
	go mod verify

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	go mod tidy
	go mod verify
	go fmt ./...
	go vet ./...
	staticcheck ./...
	go test ./... --vet --cover --short --race

# tests: run tests
.PHONY: tests
tests:
	go test ./... -v --cover --short --race

# coverage: Inspect coverage
.PHONY: coverage
coverage:
	go test -v -coverprofile cover.out ./...
	go tool cover -html cover.out -o cover.html
	open cover.html

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	cd app & \
	go build -ldflags=${LINKER_FLAGS} -o=./banking-api
