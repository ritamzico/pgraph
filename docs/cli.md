# CLI Application

The CLI provides an interactive REPL for managing multiple named graphs and running DSL queries.

```bash
make run-cli
# or
./bin/pgraph-cli
```

## REPL Commands

| Command | Description |
|---|---|
| `new <name>` | Create a new empty graph |
| `load <name> <file>` | Load a graph from a JSON file |
| `unload <name>` | Remove a loaded graph |
| `list` | List all loaded graphs (active graph marked with `*`) |
| `use <name>` | Set the active graph for queries |
| `help` | Show help |
| `exit` / `quit` | Exit the REPL |

Any other input is parsed as a [DSL query](dsl.md) and executed against the active graph.

## Example Session

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
