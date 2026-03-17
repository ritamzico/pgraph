# pgraph

A probabilistic graph inference engine written in Go. Models directed graphs where edges are independent Bernoulli random variables with associated probabilities. Designed for risk analysis, network reliability, and probabilistic reachability queries.

> **Note:** pgraph is experimental and not intended for production use. APIs, DSL syntax, and serialization formats may change without notice.

## Features

- Probabilistic directed graph model with edge probabilities
- Path inference — most probable path or top-K paths
- Exact and Monte Carlo reachability analysis
- Composite queries — AND, OR, conditional, threshold, aggregate
- Query DSL for graph construction and inference
- Go library and interactive CLI interfaces
- Batch scripting — run `.pgraph` script files for reproducible analysis

## Installation

```bash
go install github.com/ritamzico/pgraph/cmd/cli@latest        # CLI
```

Or build from source:

```bash
git clone https://github.com/ritamzico/pgraph.git
cd pgraph
make build        # produces ./bin/pgraph-cli
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

## Batch Scripting

Write a `.pgraph` script file and run it for reproducible analysis:

```pgraph
# analysis.pgraph
new supply_chain
CREATE NODE supplier, factory, retailer
CREATE EDGE e1 FROM supplier TO factory PROB 0.95
CREATE EDGE e2 FROM factory TO retailer PROB 0.9
REACHABILITY FROM supplier TO retailer EXACT
CONDITIONAL GIVEN EDGE e1 INACTIVE ( REACHABILITY FROM supplier TO retailer EXACT )
```

```bash
./bin/pgraph-cli run analysis.pgraph           # human-readable output
./bin/pgraph-cli run analysis.pgraph --json    # newline-delimited JSON
./bin/pgraph-cli run analysis.pgraph --continue  # don't stop on first error
```

## Documentation

- [Go Library API](docs/api.md)
- [CLI & Batch Mode](docs/cli.md)
- [DSL Reference](docs/dsl.md)
- [Algorithms](docs/algorithms.md)

## License

MIT
