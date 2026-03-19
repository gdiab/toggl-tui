BINARY := toggl-tui
VERSION := 0.1.0

.PHONY: build run clean install cross

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY) $(BINARY)-*

install: build
	cp $(BINARY) $(GOPATH)/bin/ 2>/dev/null || cp $(BINARY) ~/go/bin/

cross:
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY)-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY)-darwin-amd64 .
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build -o $(BINARY)-windows-amd64.exe .

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...
