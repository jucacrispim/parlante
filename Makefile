GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test -v ./...
BIN_NAME=parlante
BUILD_DIR=build
BIN_PATH=./$(BUILD_DIR)/$(BIN_NAME)
OUTFLAG=-o $(BIN_PATH)
MIGRATIONS_DIR=./migrations/

SCRIPTS_DIR=./scripts


.PHONY: build # - Creates the binary under the build/ directory
build:
	$(GOBUILD) $(OUTFLAG)

.PHONY: test # - Run all tests
test:
	$(GOBUILD)
	$(GOTEST)

.PHONY: setupenv # - Install needed tools for tests/docs
setupenv:
	$(SCRIPTS_DIR)/env.sh setup-env

.PHONY: docs # - Build documentation
docs:
	$(SCRIPTS_DIR)/env.sh build-docs

.PHONY: coverage # - Run all tests and check coverage
cov:

	$(SCRIPTS_DIR)/check_coverage.sh

coverage: cov


.PHONY: run # - Run the program. You can use `make run ARGS="-host :9090 -root=/"`
run:
	$(GOBUILD) $(OUTFLAG)
	$(BIN_PATH) $(ARGS)

.PHONY: clean # - Remove the files created during build
clean:
	rm -rf $(BUILD_DIR)

.PHONY: install # - Copy the binary to the path
install: build
	go install

.PHONY: uninstall # - Remove the binary from path
uninstall:
	go clean -i github.com/jucacrispim/parlante/$(BIN_NAME)

all: build test install

.PHONY: create_migration  # Creates new up and down migration files
create_migration:
	migrate create -ext=sql -dir=$(MIGRATIONS_DIR) -seq init


migrate_up:  # Runs the up migrations files
	migrate -path=$(MIGRATIONS_DIR) -database $(DB) -verbose up


migrate_down:  # Runs the down migrations files
	migrate -path=$(MIGRATIONS_DIR) -database $(DB) -verbose down 1


.PHONY: help  # - Show this help text
help:
	@grep '^.PHONY: .* #' Makefile | sed 's/\.PHONY: \(.*\) # \(.*\)/\1 \2/' | expand -t20
