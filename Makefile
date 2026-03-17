build:
	go build -o ./bin/pgraph-cli ./cmd/cli

run-cli:
	go run ./cmd/cli

run-batch:
	go run ./cmd/cli run $(FILE)

clean:
	rm -rf ./bin
