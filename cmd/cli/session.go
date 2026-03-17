package main

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	pgraph "github.com/ritamzico/pgraph"
)

// errExit is returned by processLine when an exit/quit command is encountered.
var errExit = errors.New("exit")

type graphEntry struct {
	pg         *pgraph.PGraph
	sourcePath string // empty if created via "new"
}

type sessionState struct {
	graphs  map[string]*graphEntry
	active  string
	scanner *bufio.Scanner // non-nil in interactive mode; nil in batch (auto-confirms saves)
}

func newSession() *sessionState {
	return &sessionState{
		graphs: make(map[string]*graphEntry),
	}
}

// processLine dispatches a single trimmed, non-empty line.
// Returns (queryResult, commandMessage, error).
// On success, exactly one of queryResult or commandMessage will be non-zero.
// Returns errExit when an exit/quit command is encountered.
func (s *sessionState) processLine(line string) (pgraph.Result, string, error) {
	parts := strings.Fields(line)
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "exit", "quit":
		return nil, "", errExit

	case "help":
		return nil, helpText, nil

	case "list":
		if len(s.graphs) == 0 {
			return nil, "(no graphs loaded)", nil
		}
		var sb strings.Builder
		for name := range s.graphs {
			marker := " "
			if name == s.active {
				marker = "*"
			}
			fmt.Fprintf(&sb, "  %s %s\n", marker, name)
		}
		return nil, strings.TrimRight(sb.String(), "\n"), nil

	case "new":
		if len(parts) < 2 {
			return nil, "", fmt.Errorf("usage: new <name>")
		}
		name := parts[1]
		s.graphs[name] = &graphEntry{pg: pgraph.New()}
		if s.active == "" {
			s.active = name
		}
		return nil, fmt.Sprintf("created empty graph %q", name), nil

	case "use":
		if len(parts) < 2 {
			return nil, "", fmt.Errorf("usage: use <name>")
		}
		name := parts[1]
		if _, ok := s.graphs[name]; !ok {
			return nil, "", fmt.Errorf("no graph named %q", name)
		}
		s.active = name
		return nil, fmt.Sprintf("active graph set to %q", name), nil

	case "load":
		if len(parts) < 3 {
			return nil, "", fmt.Errorf("usage: load <name> <file>")
		}
		name, path := parts[1], parts[2]
		pg, err := pgraph.LoadFile(path)
		if err != nil {
			return nil, "", fmt.Errorf("error loading %q: %w", path, err)
		}
		s.graphs[name] = &graphEntry{pg: pg, sourcePath: path}
		if s.active == "" {
			s.active = name
		}
		return nil, fmt.Sprintf("loaded %q (%d nodes)", name, len(pg.Graph.GetNodes())), nil

	case "save":
		if len(parts) < 2 {
			return nil, "", fmt.Errorf("usage: save <name> [file]")
		}
		name := parts[1]
		entry, ok := s.graphs[name]
		if !ok {
			return nil, "", fmt.Errorf("no graph named %q", name)
		}

		var savePath string
		if len(parts) >= 3 {
			savePath = parts[2]
		} else if entry.sourcePath != "" {
			if s.scanner != nil {
				// Interactive mode: prompt for overwrite confirmation
				fmt.Printf("graph %q was loaded from %q — overwrite? [y/N] ", name, entry.sourcePath)
				if !s.scanner.Scan() {
					return nil, "save cancelled", nil
				}
				answer := strings.TrimSpace(strings.ToLower(s.scanner.Text()))
				if answer != "y" && answer != "yes" {
					return nil, "save cancelled", nil
				}
			}
			// Batch mode (scanner == nil): auto-overwrite without prompting
			savePath = entry.sourcePath
		} else {
			return nil, "", fmt.Errorf("graph was created in-memory — specify a file path: save <name> <file>")
		}

		if err := entry.pg.SaveFile(savePath); err != nil {
			return nil, "", fmt.Errorf("error saving %q: %w", savePath, err)
		}
		entry.sourcePath = savePath
		return nil, fmt.Sprintf("saved %q to %s", name, savePath), nil

	case "unload":
		if len(parts) < 2 {
			return nil, "", fmt.Errorf("usage: unload <name>")
		}
		name := parts[1]
		if _, ok := s.graphs[name]; !ok {
			return nil, "", fmt.Errorf("no graph named %q", name)
		}
		delete(s.graphs, name)
		if s.active == name {
			s.active = ""
		}
		return nil, fmt.Sprintf("unloaded %q", name), nil

	default:
		// Treat as a DSL query against the active graph
		if s.active == "" {
			return nil, "", fmt.Errorf("no active graph — use 'load', 'use', or 'new' first")
		}
		res, err := s.graphs[s.active].pg.Query(line)
		if err != nil {
			return nil, "", fmt.Errorf("query error: %w", err)
		}
		return res, "", nil
	}
}
