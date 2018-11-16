CONATINERNAME=user-service
VERSION=1.0.0

all: dependencies generate test build docker-build

dependencies:
	GO111MODULE=on go mod download

test:
	GO111MODULE=on go test ./... --cover -tags=$(TEST_TAGS)

generate:
	GO111MODULE=on go generate ./...

build:
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server .

docker-build:
	docker build -t app/$(CONATINERNAME):$(VERSION) .

clean:
	rm -rf bin/*