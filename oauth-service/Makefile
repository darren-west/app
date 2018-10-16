CONATINERNAME=oauth-service
VERSION=1.0.0

all: dependencies generate test build docker-build

dependencies:
	go mod download

test:
	go test ./... --cover

generate:
	go generate ./...

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server .

docker-build:
	docker build -t app/$(CONATINERNAME):$(VERSION) .

clean:
	rm -rf bin/*