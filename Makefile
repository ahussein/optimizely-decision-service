### --------------------------------------------------------------------------------------------------------------------
### Variables
### (https://www.gnu.org/software/make/manual/html_node/Using-Variables.html#Using-Variables)
### --------------------------------------------------------------------------------------------------------------------
BINARY_NAME?=optimzely-decision-service
BUILD_SRC=./cmd
SRC_DIRS=internal cmd

BUILD_DIR?= build
GO_LINKER_FLAGS=-ldflags="-s -w"
GIT_HOOKS_DIR=.githooks

# colors
NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m


### --------------------------------------------------------------------------------------------------------------------
### RULES
### (https://www.gnu.org/software/make/manual/html_node/Rule-Introduction.html#Rule-Introduction)
### --------------------------------------------------------------------------------------------------------------------

# Define phony targets (https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html)
.PHONY: all clean build test-unit dev-up api-doc code-style

all: clean build


# Build the project
build: build-api build-grpc

build-api:
	@printf "$(OK_COLOR)==> Building API binary$(NO_COLOR)\n"
	@if [ ! -d ${BUILD_DIR} ] ; then mkdir -p ${BUILD_DIR} ; fi
	@GO111MODULE=on go build -o ${BUILD_DIR}/${BINARY_NAME}-api ${GO_LINKER_FLAGS} ${BUILD_SRC}/api

build-grpc:
	@printf "$(OK_COLOR)==> Building gRPC binary$(NO_COLOR)\n"
	@if [ ! -d ${BUILD_DIR} ] ; then mkdir -p ${BUILD_DIR} ; fi
	@GO111MODULE=on go build -o ${BUILD_DIR}/${BINARY_NAME}-grpc ${GO_LINKER_FLAGS} ${BUILD_SRC}/grpc

# Test the project
test-unit:
	@printf "$(OK_COLOR)==> Running tests$(NO_COLOR)\n"
	@go test -v -race -tags unit -coverprofile=coverage.txt -covermode=atomic $$(for d in $(SRC_DIRS); do echo ./$$d/...; done)

test-integration:
	@printf "$(OK_COLOR)==> Running Integration Tests$(NO_COLOR)\n"
	@go test -v -race -tags integration -coverprofile=coverage.txt -covermode=atomic $$(for d in $(SRC_DIRS); do echo ./$$d/...; done)

# Benchmark the project
test-bench:
	@printf "$(OK_COLOR)==> Running benchmarks$(NO_COLOR)\n"
	@CGO_ENABLED=0 go test -bench=. -run=XXX -v $$(for d in $(SRC_DIRS); do echo ./$$d/...; done)

# Clean after build
clean:
	@printf "$(OK_COLOR)==> Cleaning project$(NO_COLOR)\n"
	@if [ -d ${BUILD_DIR} ] ; then rm -rf ${BUILD_DIR}/* ; fi

# Set up for local development
dev-up:
	@docker-compose up --remove-orphans -d

# Runs code style checker
code-style:
ifeq ("$(command -v gometalinter >/dev/null 2>&1)","")
	@printf "$(OK_COLOR)==> Running gometalinter$(NO_COLOR)\n"
	gometalinter --disable-all --enable=vet --enable=golint --enable=goimports --deadline=120s --vendor ./...
else
	@printf "$(OK_COLOR)==> Gometalinter not present
endif

setup-hooks:
	@git config core.hooksPath $(GIT_HOOKS_DIR)
	@git config hooks.gitleaks true
	@find $(GIT_HOOKS_DIR) -type f -exec chmod 775 {} \;