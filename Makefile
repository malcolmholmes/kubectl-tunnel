.PHONY: build test install clean

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=kubectl-tunnel

# Build the project
build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/kubectl-tunnel

# Install the binary
install:
	$(GOGET) ./...
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/kubectl-tunnel
	sudo mv $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
