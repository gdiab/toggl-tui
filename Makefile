BINARY := toggl-tui
VERSION := 0.1.0
GOBIN ?= $(shell go env GOPATH)/bin

.PHONY: build run clean install uninstall cross test fmt vet

build:
	go build -ldflags "-s -w" -o $(BINARY) .

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY) $(BINARY)-*

install: build
	@mkdir -p $(GOBIN)
	cp $(BINARY) $(GOBIN)/$(BINARY)
	@echo "Installed to $(GOBIN)/$(BINARY)"
	@echo "Make sure $(GOBIN) is in your PATH"

uninstall:
	rm -f $(GOBIN)/$(BINARY)
	@echo "Removed $(GOBIN)/$(BINARY)"

cross:
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o $(BINARY)-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o $(BINARY)-darwin-amd64 .
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o $(BINARY)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o $(BINARY)-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $(BINARY)-windows-amd64.exe .

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...
