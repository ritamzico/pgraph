# HTTP Server

The server exposes a single stateless REST endpoint for executing DSL queries. It holds no state — the client is responsible for persisting graphs (e.g. in `localStorage`) and supplying the graph with every request.

```bash
make run-server
# or
./bin/pgraph-server -port 8080
```

## Endpoint

### `POST /query`

Deserialize a graph, execute a [DSL query](dsl.md) against it, and return the result. The graph is discarded immediately after the response.

**Request body:**
```json
{
  "graph": { ...graph JSON... },
  "dsl": "REACHABILITY FROM supplier TO retailer EXACT"
}
```

**Response** (`200 OK`):
```json
{"kind": "probability", "data": {"Probability": 0.855}}
```

Both fields are required. Missing or malformed input returns `400 Bad Request`. DSL errors return `422 Unprocessable Entity`.

## Graph JSON Format

The `graph` field uses the standard pgraph serialization format:

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

## CORS

Allowed origins are hardcoded in `cmd/server/main.go` in the `allowedOrigins` variable. Add new origins there as needed.

## Example

```bash
# Start the server
pgraph-server -port 8080

# Execute a query, supplying the graph inline
curl -X POST http://localhost:8080/query \
  -H 'Content-Type: application/json' \
  -d '{
    "graph": {
      "nodes": [{"id": "a", "props": {}}, {"id": "b", "props": {}}],
      "edges": [{"id": "e1", "from": "a", "to": "b", "probability": 0.9, "props": {}}]
    },
    "dsl": "REACHABILITY FROM a TO b EXACT"
  }'
```
