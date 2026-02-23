# HTTP Server

The server exposes a REST API for managing graphs and executing queries.

```bash
make run-server
# or
./bin/pgraph-server -port 8080
```

## Endpoints

### `GET /graphs`

List all loaded graph names.

**Response:**
```json
{"graphs": ["myGraph", "supplyChain"]}
```

### `POST /graphs/{name}`

Load a graph from a JSON body. Creates or replaces the named graph.

**Request body:** Graph JSON (see [Graph JSON Format](#graph-json-format))

**Response** (`201 Created`):
```json
{"name": "myGraph", "nodes": 4, "edges": 3}
```

### `DELETE /graphs/{name}`

Unload a named graph.

**Response:** `204 No Content`

### `POST /graphs/{name}/query`

Execute a [DSL query](dsl.md) against a named graph.

**Request body:**
```json
{"dsl": "REACHABILITY FROM supplier TO retailer EXACT"}
```

**Response** (`200 OK`):
```json
{"kind": "probability", "data": {"Probability": 0.855}}
```

## Graph JSON Format

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

## Example

```bash
# Start the server
pgraph-server -port 8080

# Load a graph from JSON
curl -X POST http://localhost:8080/graphs/myGraph \
  -H 'Content-Type: application/json' \
  -d @graph.json

# Run a query
curl -X POST http://localhost:8080/graphs/myGraph/query \
  -H 'Content-Type: application/json' \
  -d '{"dsl": "REACHABILITY FROM supplier TO retailer EXACT"}'
```
