
all: generate build test

test:
	go test ./... --cover

generate:
	go generate ./...

build:
	go build -o bin/server
#	go build -buildmode=plugin -o bin/google-provider.so provider/google/google.go

clean:
	rm -rf bin/*