
all: generate build test

test:
	go test ./... --cover

generate:
	go generate ./...

build:
	go build -o bin/server

clean:
	rm -rf bin/*