DOCKER_IMAGE ?= pgraph-server
DOCKER_PORT  ?= 8080

build: build-cli build-server

build-cli:
	go build -o ./bin/pgraph-cli ./cmd/cli

build-server:
	go build -o ./bin/pgraph-server ./cmd/server

run-cli:
	go run ./cmd/cli/main.go

run-server:
	go run ./cmd/server/main.go

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	docker run -d --rm --name $(DOCKER_IMAGE) -p $(DOCKER_PORT):8080 $(DOCKER_IMAGE)

clean:
	rm -rf ./bin
