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
	@go build -mod=readonly -o build/cvp ./cmd/cvp
	@echo "Compiled successfully, the output binary is located in build/"
.PHONY: build

###############################################################################
###                                 Install                                 ###
###############################################################################

install: go.sum
	@echo "Installing binary"
	@go install -mod=readonly ./cmd/cvp
	@echo "Installed successfully"
.PHONY: install