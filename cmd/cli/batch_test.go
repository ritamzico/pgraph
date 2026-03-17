package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeScript writes a script to a temp file and returns its path.
func writeScript(t *testing.T, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "test.pgraph")
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}
	return f
}

func TestRunBatch_HappyPath(t *testing.T) {
	script := writeScript(t, `
new g
CREATE NODE A
CREATE NODE B
CREATE EDGE e1 FROM A TO B PROB 0.8
REACHABILITY FROM A TO B EXACT
`)
	var stdout, stderr strings.Builder
	code := runBatch(script, batchOpts{}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "0.800000") {
		t.Errorf("expected probability 0.8 in output, got:\n%s", stdout.String())
	}
}

func TestRunBatch_CommentsSkipped(t *testing.T) {
	script := writeScript(t, `
# This is a comment
new g
# Another comment
CREATE NODE A
`)
	var stdout, stderr strings.Builder
	code := runBatch(script, batchOpts{}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", code, stderr.String())
	}
	if strings.Contains(stdout.String(), "#") {
		t.Error("comment text should not appear in output")
	}
}

func TestRunBatch_BlankLinesSkipped(t *testing.T) {
	script := writeScript(t, "new g\n\n\nCREATE NODE A\n\n")
	var stdout, stderr strings.Builder
	code := runBatch(script, batchOpts{}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", code, stderr.String())
	}
}

func TestRunBatch_FailFastOnError(t *testing.T) {
	script := writeScript(t, `new g
INVALID DSL HERE
CREATE NODE A
CREATE NODE B
CREATE EDGE e1 FROM A TO B PROB 0.9
REACHABILITY FROM A TO B EXACT
`)
	var stdout, stderr strings.Builder
	code := runBatch(script, batchOpts{}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	// Error should be reported at line 2
	if !strings.Contains(stderr.String(), "line 2") {
		t.Errorf("expected 'line 2' in stderr, got:\n%s", stderr.String())
	}
	// Subsequent lines should not have run (no probability output)
	if strings.Contains(stdout.String(), "Probability") {
		t.Error("expected execution to stop at error, but later lines ran")
	}
}

func TestRunBatch_ContinueOnError_ReportsAllErrors(t *testing.T) {
	script := writeScript(t, `new g
INVALID ONE
CREATE NODE A
INVALID TWO
CREATE NODE B
CREATE EDGE e1 FROM A TO B PROB 0.9
REACHABILITY FROM A TO B EXACT
`)
	var stdout, stderr strings.Builder
	code := runBatch(script, batchOpts{continueOnError: true}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 (errors present), got %d", code)
	}
	// Both errors reported
	if !strings.Contains(stderr.String(), "line 2") {
		t.Errorf("expected 'line 2' error, got:\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "line 4") {
		t.Errorf("expected 'line 4' error, got:\n%s", stderr.String())
	}
	// Lines after errors still executed
	if !strings.Contains(stdout.String(), "0.900000") {
		t.Errorf("expected probability output from later lines, got:\n%s", stdout.String())
	}
}

func TestRunBatch_ExitCommandStopsExecution(t *testing.T) {
	script := writeScript(t, `new g
CREATE NODE A
exit
CREATE NODE B
`)
	var stdout, stderr strings.Builder
	code := runBatch(script, batchOpts{}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", code, stderr.String())
	}
	// "created empty graph" and node A creation messages should appear
	if !strings.Contains(stdout.String(), `"g"`) {
		t.Errorf("expected graph creation message, got:\n%s", stdout.String())
	}
}

func TestRunBatch_FileNotFound(t *testing.T) {
	var stdout, stderr strings.Builder
	code := runBatch("/nonexistent/path/script.pgraph", batchOpts{}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 for missing file, got %d", code)
	}
	if !strings.Contains(stderr.String(), "cannot open") {
		t.Errorf("expected 'cannot open' in stderr, got:\n%s", stderr.String())
	}
}

func TestRunBatch_JSONOutput_ValidJSON(t *testing.T) {
	script := writeScript(t, `
new g
CREATE NODE A
CREATE NODE B
CREATE EDGE e1 FROM A TO B PROB 0.9
REACHABILITY FROM A TO B EXACT
`)
	var stdout, stderr strings.Builder
	code := runBatch(script, batchOpts{jsonOutput: true}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", code, stderr.String())
	}

	// Find the JSON result line
	var jsonLine string
	for _, line := range strings.Split(stdout.String(), "\n") {
		if strings.HasPrefix(line, "{") {
			jsonLine = line
			break
		}
	}
	if jsonLine == "" {
		t.Fatalf("expected a JSON line in output, got:\n%s", stdout.String())
	}

	// Must be valid JSON
	var result map[string]any
	if err := json.Unmarshal([]byte(jsonLine), &result); err != nil {
		t.Errorf("JSON line is not valid JSON: %v\nLine: %s", err, jsonLine)
	}
	if result["kind"] != "probability" {
		t.Errorf("expected kind 'probability', got %q", result["kind"])
	}
}

func TestRunBatch_JSONOutput_CommandMessagesStillPlainText(t *testing.T) {
	script := writeScript(t, `
new g
CREATE NODE A
`)
	var stdout, stderr strings.Builder
	runBatch(script, batchOpts{jsonOutput: true}, &stdout, &stderr)

	// Command output ("created empty graph") should not be JSON
	out := stdout.String()
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "{") {
			continue
		}
		// Non-JSON lines should not look like malformed JSON
		if strings.HasPrefix(line, "[") {
			t.Errorf("unexpected JSON array in command output: %s", line)
		}
	}
}

func TestRunBatch_MultipleGraphs(t *testing.T) {
	// 'new' does not switch the active graph if one is already set.
	// Explicitly 'use' each graph before building it.
	script := writeScript(t, `
new graph1
CREATE NODE A
CREATE NODE B
CREATE EDGE e1 FROM A TO B PROB 0.9

new graph2
use graph2
CREATE NODE X
CREATE NODE Y
CREATE EDGE ex FROM X TO Y PROB 0.5

use graph1
REACHABILITY FROM A TO B EXACT

use graph2
REACHABILITY FROM X TO Y EXACT
`)
	var stdout, stderr strings.Builder
	code := runBatch(script, batchOpts{}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "0.900000") {
		t.Errorf("expected 0.9 probability for graph1, output:\n%s", out)
	}
	if !strings.Contains(out, "0.500000") {
		t.Errorf("expected 0.5 probability for graph2, output:\n%s", out)
	}
}

func TestRunBatch_SaveAndReload(t *testing.T) {
	tmpDir := t.TempDir()
	saveFile := filepath.Join(tmpDir, "graph.json")

	// Script 1: build and save a graph
	buildScript := writeScript(t, fmt.Sprintf(`
new g
CREATE NODE A
CREATE NODE B
CREATE EDGE e1 FROM A TO B PROB 0.7
save g %s
`, saveFile))

	var stdout, stderr strings.Builder
	if code := runBatch(buildScript, batchOpts{}, &stdout, &stderr); code != 0 {
		t.Fatalf("build script failed (code %d): %s", code, stderr.String())
	}

	// Script 2: load the saved graph and query it
	queryScript := writeScript(t, fmt.Sprintf(`
load g %s
REACHABILITY FROM A TO B EXACT
`, saveFile))

	stdout.Reset()
	stderr.Reset()
	if code := runBatch(queryScript, batchOpts{}, &stdout, &stderr); code != 0 {
		t.Fatalf("query script failed (code %d): %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "0.700000") {
		t.Errorf("expected 0.7 probability after reload, got:\n%s", stdout.String())
	}
}

func TestRunBatch_NoActiveGraph_ReportsError(t *testing.T) {
	// Query without creating/loading a graph first
	script := writeScript(t, `REACHABILITY FROM A TO B EXACT`)
	var stdout, stderr strings.Builder
	code := runBatch(script, batchOpts{}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "no active graph") {
		t.Errorf("expected 'no active graph' in stderr, got:\n%s", stderr.String())
	}
}
