# pgraph

A probabilistic graph inference engine written in Go. Models directed graphs where edges are independent Bernoulli random variables with associated probabilities. Designed for risk analysis, network reliability, and other probabilistic reachability queries.

> **Note:** pgraph is an experimental project under active development. It is not intended for commercial or production use in its current state. APIs, the DSL syntax, and serialization formats may change without notice.

## Features

- **Probabilistic graph model** — directed graphs with edge probabilities representing independent failure/success events
- **Path inference** — find the most probable path or top-K most probable paths between nodes
- **Reachability analysis** — compute exact or Monte Carlo reachability probabilities
- **Composite queries** — combine queries with AND, OR, conditional, threshold, and aggregation operators
- **Query DSL** — a dedicated query language for graph construction and probabilistic inference
- **Multiple interfaces** — use pgraph as a Go library, interactive CLI, or HTTP server

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

### Go Library

```go
package main

import (
    "fmt"
    pgraph "github.com/ritamzico/pgraph"
)

func main() {
    pg := pgraph.New()

    // Build a graph
    pg.Query("CREATE NODE supplier, factory, warehouse, retailer")
    pg.Query("CREATE EDGE e1 FROM supplier TO factory PROB 0.95")
    pg.Query("CREATE EDGE e2 FROM factory TO warehouse PROB 0.9")
    pg.Query("CREATE EDGE e3 FROM warehouse TO retailer PROB 0.85")

    // Find the most probable path
    result, _ := pg.Query("MAXPATH FROM supplier TO retailer")
    fmt.Println(result)

    // Compute exact reachability probability
    result, _ = pg.Query("REACHABILITY FROM supplier TO retailer EXACT")
    fmt.Println(result)
}
```

### CLI

```
$ pgraph-cli
pgraph — probabilistic graph inference engine
Type "help" for available commands.

> new supply_chain
created empty graph "supply_chain"
[supply_chain]> CREATE NODE supplier, factory, retailer
[supply_chain]> CREATE EDGE e1 FROM supplier TO factory PROB 0.95
[supply_chain]> CREATE EDGE e2 FROM factory TO retailer PROB 0.9
[supply_chain]> REACHABILITY FROM supplier TO retailer EXACT
Reachability Probability: 0.855000
```

### HTTP Server

```bash
$ pgraph-server -port 8080
pgraph server listening on :8080
```

```bash
# Load a graph from JSON
curl -X POST http://localhost:8080/graphs/myGraph \
  -H 'Content-Type: application/json' \
  -d @graph.json

# Run a query
curl -X POST http://localhost:8080/graphs/myGraph/query \
  -H 'Content-Type: application/json' \
  -d '{"dsl": "REACHABILITY FROM supplier TO retailer EXACT"}'
```

---

## Go Library API

Import the package:

```go
import pgraph "github.com/ritamzico/pgraph"
```

### Creating and Loading Graphs

```go
// Create a new empty graph
pg := pgraph.New()

// Load from a JSON file
pg, err := pgraph.LoadFile("graph.json")

// Load from an io.Reader
pg, err := pgraph.Load(reader)
```

### Querying

All graph operations (creating nodes/edges, running queries) go through the `Query` method, which accepts the DSL syntax documented below.

```go
result, err := pg.Query("MAXPATH FROM a TO b")
```

The returned `Result` is an interface. Use type assertions to access specific result types:

```go
switch r := result.(type) {
case pgraph.PathResult:
    fmt.Printf("Path: %v (probability: %.4f)\n", r.Path.NodeIDs, r.Path.Probability)
case pgraph.PathsResult:
    for _, p := range r.Paths {
        fmt.Printf("  %v (%.4f)\n", p.NodeIDs, p.Probability)
    }
case pgraph.ProbabilityResult:
    fmt.Printf("Probability: %.6f\n", r.Probability)
case pgraph.SampleResult:
    fmt.Printf("Estimate: %.6f [%.6f, %.6f] (95%% CI)\n", r.Estimate, r.CI95Low, r.CI95High)
case pgraph.BooleanResult:
    fmt.Printf("Result: %v\n", r.Value)
case pgraph.MultiResult:
    for _, sub := range r.Results {
        fmt.Println(sub)
    }
}
```

### Saving Graphs

```go
// Save to a JSON file
err := pg.SaveFile("graph.json")

// Save to an io.Writer
err := pg.Save(writer)
```

### JSON Result Marshaling

For serializing query results (useful when building services on top of pgraph):

```go
jsonBytes, err := pgraph.MarshalResultJSON(result)
```

Each result is marshaled as `{"kind": "<type>", "data": ...}` where `kind` is one of: `path`, `paths`, `probability`, `sample`, `boolean`, `multi`.

---

## CLI Application

The CLI provides an interactive REPL for managing multiple named graphs and running DSL queries.

```bash
make run-cli
# or
./bin/pgraph-cli
```

### REPL Commands

| Command | Description |
|---|---|
| `new <name>` | Create a new empty graph |
| `load <name> <file>` | Load a graph from a JSON file |
| `unload <name>` | Remove a loaded graph |
| `list` | List all loaded graphs (active graph marked with `*`) |
| `use <name>` | Set the active graph for queries |
| `help` | Show help |
| `exit` / `quit` | Exit the REPL |

Any other input is parsed as a DSL query and executed against the active graph.

---

## HTTP Server

The server exposes a REST API for managing graphs and executing queries.

```bash
make run-server
# or
./bin/pgraph-server -port 8080
```

### Endpoints

#### `GET /graphs`

List all loaded graph names.

**Response:**
```json
{"graphs": ["myGraph", "supplyChain"]}
```

#### `POST /graphs/{name}`

Load a graph from a JSON body. Creates or replaces the named graph.

**Request body:** Graph JSON (see [Graph JSON Format](#graph-json-format))

**Response** (`201 Created`):
```json
{"name": "myGraph", "nodes": 4, "edges": 3}
```

#### `DELETE /graphs/{name}`

Unload a named graph.

**Response:** `204 No Content`

#### `POST /graphs/{name}/query`

Execute a DSL query against a named graph.

**Request body:**
```json
{"dsl": "REACHABILITY FROM supplier TO retailer EXACT"}
```

**Response** (`200 OK`):
```json
{"kind": "probability", "data": {"Probability": 0.855}}
```

### Graph JSON Format

Graphs are serialized as JSON with the following structure:

```json
{
  "nodes": [
    {
      "id": "supplier",
      "props": {
        "region": {"kind": "string", "value": "US"}
      }
    }
  ],
  "edges": [
    {
      "id": "e1",
      "from": "supplier",
      "to": "factory",
      "probability": 0.95,
      "props": {}
    }
  ]
}
```

Property values use a tagged format where `kind` is one of `int`, `float`, `string`, or `bool`.

---

## DSL Reference

pgraph includes a query DSL (Domain-Specific Language) for graph construction and probabilistic inference.

**Case sensitivity:** Keywords (`CREATE`, `NODE`, `FROM`, `REACHABILITY`, `TRUE`, etc.) are case-insensitive — `create node X`, `CREATE NODE X`, and `CrEaTe NoDe X` are all equivalent. Node names, edge names, and property keys are case-sensitive — `NodeA` and `nodea` are distinct identifiers.

**Identifiers:** Node IDs, edge IDs, and property keys must start with a letter or underscore and contain only letters, digits, and underscores (regex: `[a-zA-Z_][a-zA-Z0-9_]*`). DSL keywords (`create`, `from`, `true`, `edge`, etc.) are reserved and cannot be used as identifiers.

### Graph Construction

#### CREATE NODE

Create one or more nodes, with optional properties.

```
CREATE NODE <id>
CREATE NODE <id1>, <id2>, <id3>
CREATE NODE <id> { <key>: <value>, ... }
CREATE NODE <id1>, <id2> { <key>: <value>, ... }
```

When properties are specified on a multi-node CREATE, the same properties are applied to all nodes.

**Property values** can be strings (`"text"`), floats (`0.85`), integers (`42`), or booleans (`true` / `false`).

**Examples:**
```
CREATE NODE supplier
CREATE NODE factoryA, factoryB, warehouse, retailer
CREATE NODE supplier { region: "US", risk_score: 0.85, priority: 1, is_active: true }
CREATE NODE warehouseA, warehouseB { type: "regional" }
```

#### CREATE EDGE

Create a directed edge between two existing nodes with an associated probability, and optional properties.

```
CREATE EDGE <edgeId> FROM <sourceNode> TO <targetNode> PROB <probability>
CREATE EDGE <edgeId> FROM <sourceNode> TO <targetNode> PROB <probability> { <key>: <value>, ... }
```

**Examples:**
```
CREATE EDGE e1 FROM supplier TO factory PROB 0.95
CREATE EDGE transport_link FROM factory TO warehouse PROB 0.8 { distance: 500, mode: "rail" }
```

The probability must be a float between 0.0 and 1.0. It represents the independent probability that this edge is "active" (i.e., the connection succeeds).

#### DELETE NODE

Remove one or more nodes and all their incident edges.

```
DELETE NODE <id>
DELETE NODE <id1>, <id2>
```

#### DELETE EDGE

Remove an edge by its source and target nodes, or by its edge ID.

```
DELETE EDGE FROM <sourceNode> TO <targetNode>
DELETE EDGE <edgeId>
```

**Examples:**
```
DELETE EDGE FROM supplier TO factory
DELETE EDGE e1
```

---

### Simple Queries

#### MAXPATH

Find the single most probable path between two nodes. Uses a modified Dijkstra's algorithm with `-log(probability)` as edge weights, converting probability maximization into shortest-path minimization.

```
MAXPATH FROM <source> TO <target>
```

**Returns:** `PathResult` — the path (sequence of node IDs) and its joint probability (product of edge probabilities along the path).

**Example:**
```
MAXPATH FROM supplier TO retailer
```

#### TOPK

Find the top K most probable paths between two nodes. Uses Yen's K-shortest paths algorithm adapted for probabilistic graphs.

```
TOPK FROM <source> TO <target> K <count>
```

**Returns:** `PathsResult` — up to K paths, each with its probability.

**Example:**
```
TOPK FROM supplier TO retailer K 5
```

#### REACHABILITY (Exact)

Compute the exact probability that a target node is reachable from a source node, considering all possible paths and their edge probabilities. Uses DFS with memoization.

The formula at each node is:

```
P(current -> target) = 1 - product(1 - P(edge) * P(child -> target))
```

This accounts for multiple independent paths: the probability of reaching the target is one minus the probability of failing via every outgoing edge.

```
REACHABILITY FROM <source> TO <target> EXACT
```

**Returns:** `ProbabilityResult` — exact reachability probability.

**Example:**
```
REACHABILITY FROM supplier TO retailer EXACT
```

#### REACHABILITY (Monte Carlo)

Estimate reachability probability using parallel Monte Carlo sampling. Each sample generates a random "world" by independently activating each edge according to its probability, then tests reachability via BFS in that world.

Uses 10,000 samples distributed across worker goroutines (one per CPU core). Returns a point estimate with a 95% confidence interval.

```
REACHABILITY FROM <source> TO <target> MONTECARLO
```

**Returns:** `SampleResult` — estimated probability, number of samples, variance, standard error, and 95% CI bounds.

**Example:**
```
REACHABILITY FROM supplier TO retailer MONTECARLO
```

---

### Composite Queries

Composite queries combine multiple sub-queries. Sub-queries are separated by commas and enclosed in parentheses.

#### MULTI

Execute multiple queries in parallel and return all results.

```
MULTI ( <query1>, <query2>, ... )
```

**Returns:** `MultiResult` — a list of all sub-query results.

**Example:**
```
MULTI ( MAXPATH FROM a TO b, REACHABILITY FROM a TO b EXACT, TOPK FROM a TO b K 3 )
```

#### AND

Execute multiple probabilistic queries and combine their results assuming independence. Multiplies the probabilities together.

```
AND ( <query1>, <query2>, ... )
```

**Formula:** `P(A AND B) = P(A) * P(B)`

**Returns:** `ProbabilityResult` — the joint probability.

Sub-queries must return probabilistic results (reachability queries, other AND/OR queries, etc.).

**Example:**
```
AND ( REACHABILITY FROM a TO b EXACT, REACHABILITY FROM c TO d EXACT )
```

*"What is the probability that BOTH a can reach b AND c can reach d?"*

#### OR

Execute multiple probabilistic queries and combine using the inclusion-exclusion principle for independent events.

```
OR ( <query1>, <query2>, ... )
```

**Formula:** `P(A OR B) = 1 - (1 - P(A)) * (1 - P(B))`

**Returns:** `ProbabilityResult` — the combined probability.

**Example:**
```
OR ( REACHABILITY FROM a TO z EXACT, REACHABILITY FROM b TO z EXACT )
```

*"What is the probability that z is reachable from EITHER a or b (or both)?"*

#### CONDITIONAL

Execute a query on a conditioned graph where specific edges or nodes are forced active or inactive. The original graph is not modified; a conditioned clone is created.

```
CONDITIONAL GIVEN <condition1>, <condition2>, ... ( <query> )
```

Each condition is one of:

```
EDGE <edgeId> ACTIVE
EDGE <edgeId> INACTIVE
NODE <nodeId> ACTIVE
NODE <nodeId> INACTIVE
```

- **ACTIVE** forces the edge/node to be present (edge probability treated as 1.0)
- **INACTIVE** removes the edge/node from the graph entirely

**Returns:** The result of the inner query, executed on the conditioned graph.

**Examples:**
```
CONDITIONAL GIVEN EDGE e1 INACTIVE ( REACHABILITY FROM supplier TO retailer EXACT )
```
*"If the supplier-factory link fails, what is the reachability probability?"*

```
CONDITIONAL GIVEN EDGE e1 INACTIVE, NODE backup_supplier ACTIVE ( MAXPATH FROM backup_supplier TO retailer )
```
*"If e1 fails, what's the best path from the backup supplier?"*

#### THRESHOLD

Test whether a probabilistic query result meets a minimum probability threshold.

```
THRESHOLD <value> ( <query> )
```

The threshold value must be between 0.0 and 1.0. The inner query must return a probabilistic result.

**Returns:** `BooleanResult` — `true` if the probability >= threshold, `false` otherwise.

**Examples:**
```
THRESHOLD 0.9 ( REACHABILITY FROM supplier TO retailer EXACT )
```
*"Is the reachability probability at least 90%?"*

```
THRESHOLD 0.5 ( AND ( REACHABILITY FROM a TO b EXACT, REACHABILITY FROM c TO d EXACT ) )
```

#### AGGREGATE

Execute multiple queries and reduce their results using a named reducer function. This is useful for summarizing multiple probabilistic assessments into a single value.

```
AGGREGATE <reducer> ( <query1>, <query2>, ... )
```

Available reducers:

| Reducer | Description | Returns |
|---|---|---|
| `MEAN` | Arithmetic mean of probabilities | `ProbabilityResult` |
| `MAX` | Highest probability (best-case) | `ProbabilityResult` |
| `MIN` | Lowest probability (worst-case / weakest link) | `ProbabilityResult` |
| `BESTPATH` | Path with the highest probability | `PathResult` |
| `COUNTABOVE <float>` | Fraction of results with probability >= threshold | `ProbabilityResult` |

`MEAN`, `MAX`, `MIN`, and `COUNTABOVE` require sub-queries that return probabilistic results. `BESTPATH` requires sub-queries that return path results.

**Examples:**
```
AGGREGATE MEAN ( REACHABILITY FROM a TO b EXACT, REACHABILITY FROM c TO d EXACT )
```
*"What is the average reachability across these two pairs?"*

```
AGGREGATE MIN ( REACHABILITY FROM a TO b EXACT, REACHABILITY FROM b TO c EXACT, REACHABILITY FROM c TO d EXACT )
```
*"What is the weakest link in this chain of reachability queries?"*

```
AGGREGATE COUNTABOVE 0.9 ( REACHABILITY FROM a TO b EXACT, REACHABILITY FROM a TO c EXACT, REACHABILITY FROM a TO d EXACT )
```
*"What fraction of targets are reachable from a with at least 90% probability?"*

```
AGGREGATE BESTPATH ( MAXPATH FROM a TO d, MAXPATH FROM b TO d )
```
*"Which of these two source-to-destination paths is the most probable?"*

---

### Nesting Queries

Composite queries can be nested arbitrarily:

```
AND (
  OR ( REACHABILITY FROM a TO z EXACT, REACHABILITY FROM b TO z EXACT ),
  REACHABILITY FROM c TO z EXACT
)
```

```
THRESHOLD 0.8 (
  CONDITIONAL GIVEN EDGE backup ACTIVE (
    REACHABILITY FROM source TO sink EXACT
  )
)
```

```
MULTI (
  THRESHOLD 0.9 ( REACHABILITY FROM a TO b EXACT ),
  CONDITIONAL GIVEN EDGE e1 INACTIVE ( MAXPATH FROM a TO b ),
  TOPK FROM a TO b K 3
)
```

```
THRESHOLD 0.5 (
  AGGREGATE MEAN (
    REACHABILITY FROM a TO z EXACT,
    REACHABILITY FROM b TO z EXACT
  )
)
```

---

### DSL Grammar Summary

```
statement  = create | delete
create     = "CREATE" ("NODE" id_list props? | "EDGE" id "FROM" id "TO" id "PROB" float props?)
delete     = "DELETE" ("NODE" id_list | "EDGE" ("FROM" id "TO" id | id))

props      = "{" prop ("," prop)* "}"
prop       = id ":" value
value      = string | float | int | "TRUE" | "FALSE"

query      = simple_query | composite_query | conditional | threshold | aggregate
simple     = maxpath | topk | reachability
maxpath    = "MAXPATH" "FROM" id "TO" id
topk       = "TOPK" "FROM" id "TO" id "K" int
reachability = "REACHABILITY" "FROM" id "TO" id ("EXACT" | "MONTECARLO")

composite  = ("MULTI" | "AND" | "OR") "(" query_list ")"
query_list = query ("," query)*

conditional = "CONDITIONAL" "GIVEN" condition_list "(" query ")"
condition_list = condition ("," condition)*
condition  = ("EDGE" id | "NODE" id) ("ACTIVE" | "INACTIVE")

threshold  = "THRESHOLD" float "(" query ")"

aggregate  = "AGGREGATE" reducer "(" query_list ")"
reducer    = "MEAN" | "MAX" | "MIN" | "BESTPATH" | "COUNTABOVE" float

id         = [a-zA-Z_][a-zA-Z0-9_]*
id_list    = id ("," id)*
string     = '"' [^"\\]* '"'
float      = [0-9]+ "." [0-9]+
int        = [0-9]+
```

Keywords are case-insensitive. Identifiers (node IDs, edge IDs, property keys) are case-sensitive and cannot be reserved words.

---

## Algorithms

### Max Probability Path (Modified Dijkstra)

Finds the single path with the highest joint probability. The joint probability of a path is the product of its edge probabilities.

The algorithm converts this to a shortest-path problem using the identity:

```
max product(P_i) = min sum(-log(P_i))
```

This allows standard Dijkstra's algorithm to be applied with `-log(probability)` as edge weights.

**Complexity:** O((V + E) log V)

### Top-K Probability Paths (Yen's Algorithm)

Extends the max probability path to find the K most probable paths. Based on Yen's algorithm for K-shortest paths:

1. Find the best path using modified Dijkstra
2. For each previously found path, systematically generate candidate "spur" paths by removing edges that would duplicate known paths
3. Select the best candidate and repeat until K paths are found

### Exact Reachability (DFS with Memoization)

Computes the exact probability that a target is reachable from a source across all possible edge activation combinations:

```
P(s -> t) = 1 - product over children c of (1 - P(edge_sc) * P(c -> t))
```

Uses DFS with memoization to avoid recomputing sub-problems. Handles cycles by tracking visited nodes.

**Complexity:** O(V + E)

### Monte Carlo Reachability

Estimates reachability by sampling:

1. For each sample, independently activate each edge with its probability (Bernoulli trial)
2. Test reachability in the resulting deterministic graph using BFS
3. The fraction of samples where the target is reachable estimates the true probability
4. Returns a 95% confidence interval: `estimate +/- 1.96 * sqrt(p(1-p) / n)`

Uses parallel workers (one per CPU core), each with an independent PCG random number generator. Default: 10,000 samples.

---

## License

MIT
