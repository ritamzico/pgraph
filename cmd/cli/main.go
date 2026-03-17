package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

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

Batch mode:
  pgraph-cli run <script.pgraph> [--json] [--continue]
`

func main() {
	// Batch mode: pgraph-cli run <file> [--json] [--continue]
	if len(os.Args) >= 3 && strings.ToLower(os.Args[1]) == "run" {
		filename := os.Args[2]
		var opts batchOpts
		for _, arg := range os.Args[3:] {
			switch arg {
			case "--json":
				opts.jsonOutput = true
			case "--continue":
				opts.continueOnError = true
			}
		}
		os.Exit(runBatch(filename, opts, os.Stdout, os.Stderr))
	}

	// Interactive REPL
	s := newSession()
	scanner := bufio.NewScanner(os.Stdin)
	s.scanner = scanner

	fmt.Println("pgraph — probabilistic graph inference engine")
	fmt.Println(`Type "help" for available commands.`)
	fmt.Println()

	for {
		if s.active != "" {
			fmt.Printf("[%s]> ", s.active)
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

		res, msg, err := s.processLine(line)
		if err != nil {
			if errors.Is(err, errExit) {
				return
			}
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		if res != nil {
			fmt.Println(res.String())
		} else if msg != "" {
			fmt.Println(msg)
		}
	}
}
