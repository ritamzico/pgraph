# Go Library API

Import the package:

```go
import pgraph "github.com/ritamzico/pgraph"
```

## Creating and Loading Graphs

```go
// Create a new empty graph
pg := pgraph.New()

// Load from a JSON file
pg, err := pgraph.LoadFile("graph.json")

// Load from an io.Reader
pg, err := pgraph.Load(reader)
```

## Querying

All graph operations (creating nodes/edges, running queries) go through the `Query` method, which accepts the [DSL syntax](dsl.md).

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

## Saving Graphs

```go
// Save to a JSON file
err := pg.SaveFile("graph.json")

// Save to an io.Writer
err := pg.Save(writer)
```

## JSON Result Marshaling

For serializing query results (useful when building services on top of pgraph):

```go
jsonBytes, err := pgraph.MarshalResultJSON(result)
```

Each result is marshaled as `{"kind": "<type>", "data": ...}` where `kind` is one of: `path`, `paths`, `probability`, `sample`, `boolean`, `multi`.
