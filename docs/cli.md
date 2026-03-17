# CLI Application

The CLI provides two modes: an interactive REPL for exploratory analysis, and a batch mode for running script files.

```bash
make run-cli          # start interactive REPL
./bin/pgraph-cli      # start interactive REPL

make run-batch FILE=analysis.pgraph   # run a script file
./bin/pgraph-cli run analysis.pgraph  # run a script file
```

---

## Interactive REPL

### Commands

| Command | Description |
|---|---|
| `new <name>` | Create a new empty graph |
| `load <name> <file>` | Load a graph from a JSON file |
| `save <name> [file]` | Save a graph to a JSON file |
| `unload <name>` | Remove a loaded graph |
| `list` | List all loaded graphs (active graph marked with `*`) |
| `use <name>` | Set the active graph for queries |
| `help` | Show help |
| `exit` / `quit` | Exit the REPL |

Any other input is parsed as a [DSL query](dsl.md) and executed against the active graph.

### Example Session

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
Probability: 0.855000
```

---

## Batch Mode

Batch mode executes a script file line by line against a fresh session.

```bash
pgraph-cli run <script.pgraph> [--json] [--continue]
```

| Flag | Description |
|---|---|
| `--json` | Output DSL query results as newline-delimited JSON instead of human-readable text |
| `--continue` | Continue executing after errors instead of stopping at the first one |

**Exit codes:** `0` on success, `1` if any errors occurred.

### Script File Format

- Plain text, one statement or query per line
- Lines starting with `#` are comments and are ignored
- Blank lines are ignored
- All session commands (`new`, `load`, `save`, `use`, `unload`) are supported
- All DSL queries are supported
- Convention: use `.pgraph` extension (not enforced)

### Example Script

```pgraph
# supply_chain_analysis.pgraph
# Build and analyse a simple supply chain

new supply_chain

CREATE NODE mine { region: "AU" }
CREATE NODE factory { region: "CN" }
CREATE NODE warehouse { region: "EU" }
CREATE NODE retailer { region: "EU" }

CREATE EDGE e1 FROM mine TO factory PROB 0.95
CREATE EDGE e2 FROM factory TO warehouse PROB 0.90
CREATE EDGE e3 FROM warehouse TO retailer PROB 0.88

# End-to-end reachability
REACHABILITY FROM mine TO retailer EXACT

# Most probable path
MAXPATH FROM mine TO retailer

# What if the factory-to-warehouse link fails?
CONDITIONAL GIVEN EDGE e2 INACTIVE ( REACHABILITY FROM mine TO retailer EXACT )
```

Run it:

```bash
./bin/pgraph-cli run supply_chain_analysis.pgraph
```

Output:

```
created empty graph "supply_chain"
Probability: 0.749880
Path: mine -> factory -> warehouse -> retailer
Probability: 0.749880
Probability: 0.000000
```

### JSON Output

The `--json` flag outputs each DSL query result as a JSON object on its own line (newline-delimited JSON). Command messages (`new`, `load`, etc.) remain as plain text.

```bash
./bin/pgraph-cli run analysis.pgraph --json
```

```
created empty graph "supply_chain"
{"kind":"probability","data":{"Probability":0.74988}}
{"kind":"path","data":{"Path":{"NodeIDs":["mine","factory","warehouse","retailer"],"Probability":0.74988}}}
{"kind":"probability","data":{"Probability":0.0}}
```

The `kind` field identifies the result type: `path`, `paths`, `probability`, `sample`, `boolean`, or `multi`. Use this to filter with `jq`:

```bash
./bin/pgraph-cli run analysis.pgraph --json | jq 'select(.kind == "probability") | .data.Probability'
```

### Loading a Saved Graph

If a graph was previously saved with `save`, load it in a script with `load`:

```pgraph
# Run queries against a pre-built graph
load supply supply_chain.json
MAXPATH FROM mine TO retailer
REACHABILITY FROM mine TO retailer EXACT
```

### Save in Batch Mode

When `save <name>` is called without an explicit file path and the graph was loaded from a file, batch mode auto-overwrites without prompting (unlike the interactive REPL which asks for confirmation).

```pgraph
load supply supply_chain.json
CREATE NODE new_depot { region: "US" }
CREATE EDGE e4 FROM warehouse TO new_depot PROB 0.7
save supply   # overwrites supply_chain.json without prompt
```
