# pgraph

A probabilistic graph inference engine written in Go. Models directed graphs where edges are independent Bernoulli random variables with associated probabilities. Designed for risk analysis, network reliability, and probabilistic reachability queries.

> **Note:** pgraph is experimental and not intended for production use. APIs, DSL syntax, and serialization formats may change without notice.

## Features

- Probabilistic directed graph model with edge probabilities
- Path inference — most probable path or top-K paths
- Exact and Monte Carlo reachability analysis
- Composite queries — AND, OR, conditional, threshold, aggregate
- Query DSL for graph construction and inference
- Go library, interactive CLI, and HTTP server interfaces

## Installation

```bash
go install github.com/ritamzico/pgraph/cmd/cli@latest        # CLI
go install github.com/ritamzico/pgraph/cmd/server@latest     # HTTP server
```

Or build from source:

```bash
git clone https://github.com/ritamzico/pgraph.git
cd pgraph
make build        # produces ./bin/pgraph-cli and ./bin/pgraph-server
```

## Quick Start

```go
pg := pgraph.New()
pg.Query("CREATE NODE supplier, factory, warehouse, retailer")
pg.Query("CREATE EDGE e1 FROM supplier TO factory PROB 0.95")
pg.Query("CREATE EDGE e2 FROM factory TO warehouse PROB 0.9")
pg.Query("CREATE EDGE e3 FROM warehouse TO retailer PROB 0.85")

result, _ := pg.Query("REACHABILITY FROM supplier TO retailer EXACT")
fmt.Println(result)
```

## Documentation

- [Go Library API](docs/api.md)
- [CLI](docs/cli.md)
- [HTTP Server](docs/server.md)
- [DSL Reference](docs/dsl.md)
- [Algorithms](docs/algorithms.md)

## License

MIT
