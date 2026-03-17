package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	pgraph "github.com/ritamzico/pgraph"
)

type batchOpts struct {
	jsonOutput      bool
	continueOnError bool
}

// runBatch executes a .pgraph script file line by line against a fresh session.
// Each non-blank, non-comment line is dispatched as either a session command or
// a DSL query. Results are written to stdout; errors are written to stderr.
// Returns 0 on success, 1 if any errors occurred.
func runBatch(filename string, opts batchOpts, stdout, stderr io.Writer) int {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(stderr, "cannot open script %q: %v\n", filename, err)
		return 1
	}
	defer f.Close()

	s := newSession() // scanner is nil → batch mode (auto-confirms saves)
	scanner := bufio.NewScanner(f)
	lineNum := 0
	hasErrors := false

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip blank lines and comment lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		res, msg, err := s.processLine(line)
		if err != nil {
			if errors.Is(err, errExit) {
				break
			}
			fmt.Fprintf(stderr, "error (line %d): %v\n", lineNum, err)
			hasErrors = true
			if !opts.continueOnError {
				return 1
			}
			continue
		}

		if res != nil {
			if opts.jsonOutput {
				b, jerr := marshalAnnotated(lineNum, line, res)
				if jerr != nil {
					fmt.Fprintf(stderr, "error (line %d): marshalling result: %v\n", lineNum, jerr)
					hasErrors = true
				} else {
					fmt.Fprintln(stdout, string(b))
				}
			} else {
				fmt.Fprintf(stdout, "> %s\n%s\n", line, res.String())
			}
		} else if msg != "" {
			fmt.Fprintln(stdout, msg)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(stderr, "error reading script: %v\n", err)
		return 1
	}

	if hasErrors {
		return 1
	}
	return 0
}

// marshalAnnotated wraps a query result with the source line number and query
// text, producing: {"line":N,"query":"...","kind":"...","data":{...}}
func marshalAnnotated(lineNum int, query string, res pgraph.Result) ([]byte, error) {
	raw, err := pgraph.MarshalResultJSON(res)
	if err != nil {
		return nil, err
	}

	// Parse the base {"kind":..., "data":...} produced by MarshalResultJSON.
	var base map[string]json.RawMessage
	if err := json.Unmarshal(raw, &base); err != nil {
		return nil, err
	}

	out := struct {
		Line  int              `json:"line"`
		Query string           `json:"query"`
		Kind  json.RawMessage  `json:"kind"`
		Data  json.RawMessage  `json:"data"`
	}{
		Line:  lineNum,
		Query: query,
		Kind:  base["kind"],
		Data:  base["data"],
	}
	return json.Marshal(out)
}
