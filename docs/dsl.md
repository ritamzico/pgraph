# DSL Reference

pgraph includes a query DSL (Domain-Specific Language) for graph construction and probabilistic inference.

**Case sensitivity:** Keywords (`CREATE`, `NODE`, `FROM`, `REACHABILITY`, `TRUE`, etc.) are case-insensitive. Node names, edge names, and property keys are case-sensitive — `NodeA` and `nodea` are distinct identifiers.

**Identifiers:** Node IDs, edge IDs, and property keys must start with a letter or underscore and contain only letters, digits, and underscores (`[a-zA-Z_][a-zA-Z0-9_]*`). DSL keywords are reserved and cannot be used as identifiers.

---

## Graph Construction

### CREATE NODE

Create one or more nodes, with optional properties.

```
CREATE NODE <id>
CREATE NODE <id1>, <id2>, <id3>
CREATE NODE <id> { <key>: <value>, ... }
CREATE NODE <id1>, <id2> { <key>: <value>, ... }
```

When properties are specified on a multi-node CREATE, the same properties are applied to all nodes.

**Property values** can be strings (`"text"`), floats (`0.85`), integers (`42`), or booleans (`true` / `false`).

```
CREATE NODE supplier
CREATE NODE factoryA, factoryB, warehouse, retailer
CREATE NODE supplier { region: "US", risk_score: 0.85, priority: 1, is_active: true }
CREATE NODE warehouseA, warehouseB { type: "regional" }
```

### CREATE EDGE

Create a directed edge between two existing nodes with an associated probability, and optional properties.

```
CREATE EDGE <edgeId> FROM <sourceNode> TO <targetNode> PROB <probability>
CREATE EDGE <edgeId> FROM <sourceNode> TO <targetNode> PROB <probability> { <key>: <value>, ... }
```

```
CREATE EDGE e1 FROM supplier TO factory PROB 0.95
CREATE EDGE transport_link FROM factory TO warehouse PROB 0.8 { distance: 500, mode: "rail" }
```

The probability must be a float between 0.0 and 1.0. It represents the independent probability that this edge is "active" (i.e., the connection succeeds).

### DELETE NODE

Remove one or more nodes and all their incident edges.

```
DELETE NODE <id>
DELETE NODE <id1>, <id2>
```

### DELETE EDGE

Remove an edge by its source and target nodes, or by its edge ID.

```
DELETE EDGE FROM <sourceNode> TO <targetNode>
DELETE EDGE <edgeId>
```

---

## Simple Queries

### MAXPATH

Find the single most probable path between two nodes. Uses a modified Dijkstra's algorithm with `-log(probability)` as edge weights.

```
MAXPATH FROM <source> TO <target>
```

**Returns:** `PathResult` — the path (sequence of node IDs) and its joint probability.

```
MAXPATH FROM supplier TO retailer
```

### TOPK

Find the top K most probable paths between two nodes. Uses Yen's K-shortest paths algorithm adapted for probabilistic graphs.

```
TOPK FROM <source> TO <target> K <count>
```

**Returns:** `PathsResult` — up to K paths, each with its probability.

```
TOPK FROM supplier TO retailer K 5
```

### REACHABILITY (Exact)

Compute the exact probability that a target node is reachable from a source node, considering all possible paths. Uses DFS with memoization.

```
P(current -> target) = 1 - product(1 - P(edge) * P(child -> target))
```

```
REACHABILITY FROM <source> TO <target> EXACT
```

**Returns:** `ProbabilityResult` — exact reachability probability.

```
REACHABILITY FROM supplier TO retailer EXACT
```

### REACHABILITY (Monte Carlo)

Estimate reachability probability using parallel Monte Carlo sampling (10,000 samples across CPU-count workers). Returns a point estimate with a 95% confidence interval.

```
REACHABILITY FROM <source> TO <target> MONTECARLO
```

**Returns:** `SampleResult` — estimated probability, number of samples, variance, standard error, and 95% CI bounds.

```
REACHABILITY FROM supplier TO retailer MONTECARLO
```

---

## Composite Queries

Composite queries combine multiple sub-queries separated by commas in parentheses.

### MULTI

Execute multiple queries in parallel and return all results.

```
MULTI ( <query1>, <query2>, ... )
```

**Returns:** `MultiResult` — a list of all sub-query results.

```
MULTI ( MAXPATH FROM a TO b, REACHABILITY FROM a TO b EXACT, TOPK FROM a TO b K 3 )
```

### AND

Combine probabilistic queries assuming independence. Multiplies the probabilities together.

```
AND ( <query1>, <query2>, ... )
```

**Formula:** `P(A AND B) = P(A) * P(B)`

**Returns:** `ProbabilityResult` — the joint probability.

```
AND ( REACHABILITY FROM a TO b EXACT, REACHABILITY FROM c TO d EXACT )
```
*"What is the probability that BOTH a can reach b AND c can reach d?"*

### OR

Combine probabilistic queries using the inclusion-exclusion principle for independent events.

```
OR ( <query1>, <query2>, ... )
```

**Formula:** `P(A OR B) = 1 - (1 - P(A)) * (1 - P(B))`

**Returns:** `ProbabilityResult` — the combined probability.

```
OR ( REACHABILITY FROM a TO z EXACT, REACHABILITY FROM b TO z EXACT )
```
*"What is the probability that z is reachable from EITHER a or b (or both)?"*

### CONDITIONAL

Execute a query on a conditioned graph where specific edges or nodes are forced active or inactive. The original graph is not modified.

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

**Returns:** The result of the inner query on the conditioned graph.

```
CONDITIONAL GIVEN EDGE e1 INACTIVE ( REACHABILITY FROM supplier TO retailer EXACT )
```
*"If the supplier-factory link fails, what is the reachability probability?"*

```
CONDITIONAL GIVEN EDGE e1 INACTIVE, NODE backup_supplier ACTIVE ( MAXPATH FROM backup_supplier TO retailer )
```
*"If e1 fails, what's the best path from the backup supplier?"*

### THRESHOLD

Test whether a probabilistic query result meets a minimum probability threshold.

```
THRESHOLD <value> ( <query> )
```

**Returns:** `BooleanResult` — `true` if the probability >= threshold, `false` otherwise.

```
THRESHOLD 0.9 ( REACHABILITY FROM supplier TO retailer EXACT )
```
*"Is the reachability probability at least 90%?"*

### AGGREGATE

Execute multiple queries and reduce their results using a named reducer.

```
AGGREGATE <reducer> ( <query1>, <query2>, ... )
```

| Reducer | Description | Returns |
|---|---|---|
| `MEAN` | Arithmetic mean of probabilities | `ProbabilityResult` |
| `MAX` | Highest probability (best-case) | `ProbabilityResult` |
| `MIN` | Lowest probability (worst-case / weakest link) | `ProbabilityResult` |
| `BESTPATH` | Path with the highest probability | `PathResult` |
| `COUNTABOVE <float>` | Fraction of results with probability >= threshold | `ProbabilityResult` |

`MEAN`, `MAX`, `MIN`, and `COUNTABOVE` require sub-queries that return probabilistic results. `BESTPATH` requires sub-queries that return path results.

```
AGGREGATE MEAN ( REACHABILITY FROM a TO b EXACT, REACHABILITY FROM c TO d EXACT )
```
*"What is the average reachability across these two pairs?"*

```
AGGREGATE MIN ( REACHABILITY FROM a TO b EXACT, REACHABILITY FROM b TO c EXACT, REACHABILITY FROM c TO d EXACT )
```
*"What is the weakest link in this chain?"*

```
AGGREGATE COUNTABOVE 0.9 ( REACHABILITY FROM a TO b EXACT, REACHABILITY FROM a TO c EXACT, REACHABILITY FROM a TO d EXACT )
```
*"What fraction of targets are reachable from a with at least 90% probability?"*

```
AGGREGATE BESTPATH ( MAXPATH FROM a TO d, MAXPATH FROM b TO d )
```
*"Which of these two source-to-destination paths is the most probable?"*

---

## Nesting Queries

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

## Grammar Summary

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

Keywords are case-insensitive. Identifiers are case-sensitive and cannot be reserved words.
