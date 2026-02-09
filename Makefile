build: build-cli build-server

build-cli:
	go build -o ./bin/pgraph-cli ./cmd/cli

build-server:
	go build -o ./bin/pgraph-server ./cmd/server

run-cli:
	go run ./cmd/cli/main.go

run-server:
	go run ./cmd/server/main.go

clean:
	rm -rf ./bin
