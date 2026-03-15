build:
	go build -o ./bin/pgraph-cli ./cmd/cli

run-cli:
	go run ./cmd/cli/main.go

clean:
	rm -rf ./bin
