# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Probabilistic graph inference engine written in Go (1.25.5). Models directed graphs where edges are independent Bernoulli random variables with associated probabilities. Primary use case is supply chain risk analysis and probabilistic reachability.

**Status:** Experimental — not intended for commercial or production use.

Module: `github.com/ritamzico/pgraph`

## Build & Run Commands

```bash
make build          # Builds ./bin/pgraph-cli and ./bin/pgraph-server
make build-cli      # Builds CLI only
make build-server   # Builds server only
make run-cli        # Runs CLI via go run ./cmd/cli/main.go
make run-server     # Runs server via go run ./cmd/server/main.go
make clean          # Removes ./bin directory
go test ./...       # Run all tests
```

## Architecture

### Public API (`pgraph.go`)

The root package exposes the public Go library API:

- **`PGraph`** — main struct wrapping a graph and DSL parser.
- **`New()`** — create an empty graph.
- **`Load(io.Reader)` / `LoadFile(path)`** — deserialize a graph from JSON.
- **`Query(dsl string)`** — parse and execute a DSL statement or query, returns a `Result`.
- **`Save(io.Writer)` / `SaveFile(path)`** — serialize the graph to JSON.
- **`MarshalResultJSON(Result)`** — serialize a query result to tagged JSON (`{"kind": "...", "data": ...}`).
- **Result type aliases** — re-exports `PathResult`, `PathsResult`, `ProbabilityResult`, `SampleResult`, `MultiResult`, `BooleanResult` from `internal/result`.

### Package Structure

- **`cmd/cli/`** — Interactive REPL. Manages multiple named graphs (`new`, `load`, `unload`, `use`, `list`). Any unrecognized input is executed as a DSL query against the active graph.
- **`cmd/server/`** — HTTP server exposing a REST API. Endpoints: `GET /graphs` (list), `POST /graphs/{name}` (load from JSON body), `DELETE /graphs/{name}` (unload), `POST /graphs/{name}/query` (execute DSL). Uses `sync.RWMutex`-protected store for concurrent access.
- **`internal/graph/`** — Core data structures: `ProbabilisticGraphModel` interface, `ProbabilisticAdjacencyListGraph` implementation (bidirectional adjacency list with `nodeMap`, `edgeMap`, `out`, `in` maps), `Node`, `Edge`, `Path`, `Condition`, `Value`.
- **`internal/dsl/`** — DSL parser built with `participle/v2`. Defines the grammar AST (`grammar.go`), converts AST to domain objects (`convert.go`), and provides the `Parser` entry point (`parser.go`). Statement types for CREATE/DELETE are in `statement.go`.
- **`internal/query/`** — Query interface (`Execute(ctx, graph) -> Result`) with simple queries (`MaxProbabilityPathQuery`, `TopKProbabilityPathsQuery`, `ReachabilityProbabilityQuery`) and composite queries (`MultiQuery`, `AndQuery`, `OrQuery`, `ConditionalQuery`, `SequentialQuery`, `ThresholdQuery`, `AggregateQuery`). Reducers (`MeanProbabilityReducer`, `BestPathReducer`) for aggregate queries in `reducer.go`.
- **`internal/engine/`** — `InferenceEngine` orchestrates query execution against a graph with context support.
- **`internal/inference/`** — Algorithm implementations:
  - **MaxProbabilityPath**: Modified Dijkstra using `-log(prob)` as edge weights (`max_probability_path.go`).
  - **TopKMaxProbabilityPaths**: Yen's K-shortest paths variant (`top_k_max_probability_paths.go`).
  - **ReachabilityProbability (exact)**: DFS with memoization; `1 - product(1 - P(reach via child))` (`reachability_probability.go`, `graph_traversals.go`).
  - **ReachabilityProbability (Monte Carlo)**: Parallel sampling with goroutine worker pool sized to CPU count. Each worker gets its own PCG RNG. 10,000 samples. Returns estimate with 95% CI (`reachability_probability.go`).
  - **Priority queue**: Min-heap for Dijkstra (`priority_queue.go`).
- **`internal/sampling/`** — `WorldSampler` interface and `IndependentEdgeSampler` that generates boolean edge masks by sampling each edge independently via Bernoulli trials.
- **`internal/result/`** — Result types: `PathResult`, `PathsResult`, `ProbabilityResult`, `SampleResult` (with CI bounds), `MultiResult`, `BooleanResult`. `ProbabilisticResult` sub-interface for results that expose a probability value.
- **`internal/serialization/`** — JSON serialization/deserialization of graphs. Format: `{"nodes": [...], "edges": [...]}` with typed property values.

### Key Patterns

- **Interface-driven**: `ProbabilisticGraphModel` and `Query` interfaces allow swappable implementations. Graph conditioning (`ApplyCondition`) creates modified clones without mutating the original.
- **Concurrent execution**: Composite queries (`Multi`, `And`, `Or`) execute sub-queries concurrently via `executeConcurrent`. Monte Carlo sampling parallelizes across workers.
- **Context propagation**: All query execution respects `context.Context` for cancellation.
- **Session isolation**: The DSL parser clones the graph before executing, so mutations from CREATE/DELETE don't affect the base graph.
- **Custom errors per package**: `GraphError`, `SyntaxError`, `QueryError`, `InferenceError`.

### DSL Rules

- **Keywords** (`CREATE`, `NODE`, `FROM`, `TRUE`, etc.) are **case-insensitive**.
- **Identifiers** (node IDs, edge IDs, property keys) are **case-sensitive** and must match `[a-zA-Z_][a-zA-Z0-9_]*`. Reserved keywords cannot be used as identifiers.
- **Properties** are optional key-value blocks in `{ }` syntax. Values can be strings (`"text"`), floats (`0.85`), integers (`42`), or booleans (`true`/`false`).

### DSL Syntax Examples

```
CREATE NODE nodeA, nodeB
CREATE NODE supplier { region: "US", risk_score: 0.85 }
CREATE EDGE edgeAB FROM nodeA TO nodeB PROB 0.9
CREATE EDGE e1 FROM a TO b PROB 0.95 { distance: 500, mode: "rail" }
DELETE NODE nodeA
DELETE EDGE FROM nodeA TO nodeB
DELETE EDGE edgeAB
MAXPATH FROM nodeA TO nodeB
TOPK FROM nodeA TO nodeB K 4
REACHABILITY FROM nodeA TO nodeB EXACT
REACHABILITY FROM nodeA TO nodeB MONTECARLO
MULTI ( MAXPATH FROM a TO b, REACHABILITY FROM c TO d EXACT )
AND ( REACHABILITY FROM a TO b EXACT, REACHABILITY FROM c TO d EXACT )
OR ( REACHABILITY FROM a TO b EXACT, REACHABILITY FROM c TO d EXACT )
CONDITIONAL GIVEN EDGE e1 INACTIVE ( REACHABILITY FROM a TO b EXACT )
THRESHOLD 0.9 ( REACHABILITY FROM a TO b EXACT )
```

## Dependencies

- `github.com/alecthomas/participle/v2` — Parser generator used for DSL grammar definition.
- Standard library only otherwise (`context`, `container/heap`, `math/rand/v2`, `sync`, `runtime`, `net/http`, `encoding/json`).
