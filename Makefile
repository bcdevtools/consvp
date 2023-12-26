GIT_TAG 	:= $(shell echo $(shell git describe --tags || git branch --show-current) | sed 's/^v//')
COMMIT  	:= $(shell git log -1 --format='%H')
BUILD_DATE	:= $(shell date '+%Y-%m-%d')

###############################################################################
###                                Build flags                              ###
###############################################################################

LD_FLAGS = -X github.com/bcdevtools/consvp/constants.VERSION=$(GIT_TAG) \
	-X github.com/bcdevtools/consvp/constants.GIT_COMMIT_HASH=$(COMMIT)  \
	-X github.com/bcdevtools/consvp/constants.BUILD_DATE=$(BUILD_DATE)

BUILD_FLAGS := -ldflags '$(LD_FLAGS)'

###############################################################################
###                                  Test                                   ###
###############################################################################

test: go.sum
	@echo "Testing"
	@go test -v ./... -race -coverprofile=coverage.txt -covermode=atomic
.PHONY: test

###############################################################################
###                                  Build                                  ###
###############################################################################

build: go.sum
	@echo "Compiling binary"
	@echo "Flags $(BUILD_FLAGS)"
	@go build -mod=readonly $(BUILD_FLAGS) -o build/cvp ./cmd/cvp
	@echo "Compiled $(GIT_TAG) successfully, the output binary is located in build/"
.PHONY: build

###############################################################################
###                                 Install                                 ###
###############################################################################

install: go.sum
	@echo "Installing binary"
	@echo "Flags $(BUILD_FLAGS)"
	@go install -mod=readonly $(BUILD_FLAGS) ./cmd/cvp
	@echo "Installed $(GIT_TAG) successfully"
.PHONY: install