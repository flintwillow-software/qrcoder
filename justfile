# qrcoder — QR code CLI tool

default: build

build:
	go build -o qrcoder .

install:
	go install .

run data:
	go run . -data "{{data}}"

test:
	go test ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

lint:
	staticcheck ./...

clean:
	rm -f qrcoder
	go clean -cache
