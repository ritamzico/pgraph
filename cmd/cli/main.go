package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	pgraph "github.com/ritamzico/pgraph"
)

type graphEntry struct {
	pg         *pgraph.PGraph
	sourcePath string // empty if created via "new"
}

const helpText = `pgraph interactive REPL

Commands:
  new <name>           Create a new empty graph
  load <name> <file>   Load a graph from a JSON file
  save <name> [file]   Save a graph to a JSON file
  unload <name>        Remove a loaded graph
  list                 List all loaded graphs
  use <name>           Set the active graph for queries
  help                 Show this help message
  exit / quit          Exit the REPL

Any other input is treated as a DSL query against the active graph.

DSL examples:
  MAXPATH FROM nodeA TO nodeB
  TOPK FROM nodeA TO nodeB K 3
  REACHABILITY FROM nodeA TO nodeB EXACT
  REACHABILITY FROM nodeA TO nodeB MONTECARLO
  CREATE NODE myNode
  CREATE EDGE e1 FROM nodeA TO nodeB PROB 0.8
`

func main() {
	graphs := make(map[string]*graphEntry)
	var active string

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("pgraph — probabilistic graph inference engine")
	fmt.Println(`Type "help" for available commands.`)
	fmt.Println()

	for {
		if active != "" {
			fmt.Printf("[%s]> ", active)
		} else {
			fmt.Print("> ")
		}

		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "exit", "quit":
			return

		case "help":
			fmt.Print(helpText)

		case "list":
			if len(graphs) == 0 {
				fmt.Println("(no graphs loaded)")
			} else {
				for name := range graphs {
					marker := " "
					if name == active {
						marker = "*"
					}
					fmt.Printf("  %s %s\n", marker, name)
				}
			}

		case "new":
			if len(parts) < 2 {
				fmt.Fprintln(os.Stderr, "usage: new <name>")
				continue
			}
			name := parts[1]
			graphs[name] = &graphEntry{pg: pgraph.New()}
			if active == "" {
				active = name
			}
			fmt.Printf("created empty graph %q\n", name)

		case "use":
			if len(parts) < 2 {
				fmt.Fprintln(os.Stderr, "usage: use <name>")
				continue
			}
			name := parts[1]
			if _, ok := graphs[name]; !ok {
				fmt.Fprintf(os.Stderr, "no graph named %q\n", name)
				continue
			}
			active = name
			fmt.Printf("active graph set to %q\n", name)

		case "load":
			if len(parts) < 3 {
				fmt.Fprintln(os.Stderr, "usage: load <name> <file>")
				continue
			}
			name, path := parts[1], parts[2]
			pg, err := pgraph.LoadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error loading %q: %v\n", path, err)
				continue
			}
			graphs[name] = &graphEntry{pg: pg, sourcePath: path}
			if active == "" {
				active = name
			}
			fmt.Printf("loaded %q (%d nodes)\n", name, len(pg.Graph.GetNodes()))

		case "save":
			if len(parts) < 2 {
				fmt.Fprintln(os.Stderr, "usage: save <name> [file]")
				continue
			}
			name := parts[1]
			entry, ok := graphs[name]
			if !ok {
				fmt.Fprintf(os.Stderr, "no graph named %q\n", name)
				continue
			}

			var savePath string
			if len(parts) >= 3 {
				savePath = parts[2]
			} else if entry.sourcePath != "" {
				// Loaded from file — confirm overwrite
				fmt.Printf("graph %q was loaded from %q — overwrite? [y/N] ", name, entry.sourcePath)
				if !scanner.Scan() {
					break
				}
				answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
				if answer != "y" && answer != "yes" {
					fmt.Println("save cancelled")
					continue
				}
				savePath = entry.sourcePath
			} else {
				fmt.Fprintln(os.Stderr, "graph was created in-memory — specify a file path: save <name> <file>")
				continue
			}

			if err := entry.pg.SaveFile(savePath); err != nil {
				fmt.Fprintf(os.Stderr, "error saving %q: %v\n", savePath, err)
				continue
			}
			entry.sourcePath = savePath
			fmt.Printf("saved %q to %s\n", name, savePath)

		case "unload":
			if len(parts) < 2 {
				fmt.Fprintln(os.Stderr, "usage: unload <name>")
				continue
			}
			name := parts[1]
			if _, ok := graphs[name]; !ok {
				fmt.Fprintf(os.Stderr, "no graph named %q\n", name)
				continue
			}
			delete(graphs, name)
			if active == name {
				active = ""
			}
			fmt.Printf("unloaded %q\n", name)

		default:
			if active == "" {
				fmt.Fprintln(os.Stderr, "no active graph — use 'load' or 'use' first")
				continue
			}
			res, err := graphs[active].pg.Query(line)
			if err != nil {
				fmt.Fprintf(os.Stderr, "query error: %v\n", err)
			} else if res != nil {
				fmt.Println(res.String())
			}
		}
	}
}
