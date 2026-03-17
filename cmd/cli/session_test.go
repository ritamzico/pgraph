package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// probabilistic matches any result that exposes a probability value,
// matching the result.ProbabilisticResult interface without importing internal packages.
type probabilistic interface {
	ProbabilityValue() float64
}

// --- exit / quit ---

func TestProcessLine_Exit(t *testing.T) {
	s := newSession()
	_, _, err := s.processLine("exit")
	if !errors.Is(err, errExit) {
		t.Errorf("expected errExit, got %v", err)
	}
}

func TestProcessLine_Quit(t *testing.T) {
	s := newSession()
	_, _, err := s.processLine("quit")
	if !errors.Is(err, errExit) {
		t.Errorf("expected errExit, got %v", err)
	}
}

func TestProcessLine_ExitCaseInsensitive(t *testing.T) {
	for _, cmd := range []string{"EXIT", "Exit", "QUIT", "Quit"} {
		s := newSession()
		_, _, err := s.processLine(cmd)
		if !errors.Is(err, errExit) {
			t.Errorf("%q: expected errExit, got %v", cmd, err)
		}
	}
}

// --- help ---

func TestProcessLine_Help(t *testing.T) {
	s := newSession()
	_, msg, err := s.processLine("help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg == "" {
		t.Error("expected non-empty help text")
	}
}

// --- new ---

func TestProcessLine_New_CreatesGraph(t *testing.T) {
	s := newSession()
	_, msg, err := s.processLine("new mygraph")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.graphs["mygraph"]; !ok {
		t.Error("graph 'mygraph' should exist after new")
	}
	if !strings.Contains(msg, "mygraph") {
		t.Errorf("expected message to mention graph name, got %q", msg)
	}
}

func TestProcessLine_New_SetsFirstGraphActive(t *testing.T) {
	s := newSession()
	s.processLine("new first")
	if s.active != "first" {
		t.Errorf("expected 'first' to be active, got %q", s.active)
	}
	s.processLine("new second")
	if s.active != "first" {
		t.Errorf("expected 'first' to remain active after adding 'second', got %q", s.active)
	}
}

func TestProcessLine_New_MissingName(t *testing.T) {
	s := newSession()
	_, _, err := s.processLine("new")
	if err == nil {
		t.Error("expected error for 'new' with no name")
	}
}

// --- use ---

func TestProcessLine_Use_SwitchesActive(t *testing.T) {
	s := newSession()
	s.processLine("new alpha")
	s.processLine("new beta")
	_, _, err := s.processLine("use beta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.active != "beta" {
		t.Errorf("expected active 'beta', got %q", s.active)
	}
}

func TestProcessLine_Use_UnknownGraph(t *testing.T) {
	s := newSession()
	_, _, err := s.processLine("use nonexistent")
	if err == nil {
		t.Error("expected error for unknown graph")
	}
}

func TestProcessLine_Use_MissingName(t *testing.T) {
	s := newSession()
	_, _, err := s.processLine("use")
	if err == nil {
		t.Error("expected error for 'use' with no name")
	}
}

// --- list ---

func TestProcessLine_List_Empty(t *testing.T) {
	s := newSession()
	_, msg, err := s.processLine("list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "(no graphs loaded)" {
		t.Errorf("expected '(no graphs loaded)', got %q", msg)
	}
}

func TestProcessLine_List_MarksActiveGraph(t *testing.T) {
	s := newSession()
	s.processLine("new g1")
	s.processLine("new g2")
	s.processLine("use g2")
	_, msg, err := s.processLine("list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(msg, "* g2") {
		t.Errorf("expected active graph g2 marked with *, got:\n%s", msg)
	}
	if !strings.Contains(msg, "  g1") {
		t.Errorf("expected inactive graph g1 with space prefix, got:\n%s", msg)
	}
}

// --- unload ---

func TestProcessLine_Unload_RemovesGraph(t *testing.T) {
	s := newSession()
	s.processLine("new g1")
	_, _, err := s.processLine("unload g1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.graphs["g1"]; ok {
		t.Error("graph 'g1' should be removed after unload")
	}
}

func TestProcessLine_Unload_ClearsActive(t *testing.T) {
	s := newSession()
	s.processLine("new g1")
	s.processLine("unload g1")
	if s.active != "" {
		t.Errorf("expected active to be cleared, got %q", s.active)
	}
}

func TestProcessLine_Unload_DoesNotClearOtherActive(t *testing.T) {
	s := newSession()
	s.processLine("new g1")
	s.processLine("new g2")
	s.processLine("use g2")
	s.processLine("unload g1")
	if s.active != "g2" {
		t.Errorf("expected active to remain 'g2', got %q", s.active)
	}
}

func TestProcessLine_Unload_UnknownGraph(t *testing.T) {
	s := newSession()
	_, _, err := s.processLine("unload nonexistent")
	if err == nil {
		t.Error("expected error for unloading unknown graph")
	}
}

func TestProcessLine_Unload_MissingName(t *testing.T) {
	s := newSession()
	_, _, err := s.processLine("unload")
	if err == nil {
		t.Error("expected error for 'unload' with no name")
	}
}

// --- load ---

func TestProcessLine_Load_FromFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "graph.json")
	graphJSON := `{"nodes":[{"id":"X","props":null},{"id":"Y","props":null}],"edges":[{"id":"e1","from":"X","to":"Y","probability":0.9,"props":null}]}`
	if err := os.WriteFile(tmpFile, []byte(graphJSON), 0644); err != nil {
		t.Fatalf("failed to write graph file: %v", err)
	}

	s := newSession()
	_, msg, err := s.processLine("load g " + tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.graphs["g"]; !ok {
		t.Error("graph 'g' should exist after load")
	}
	if !strings.Contains(msg, "2") {
		t.Errorf("expected node count (2) in message, got %q", msg)
	}
}

func TestProcessLine_Load_SetsFirstGraphActive(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "graph.json")
	os.WriteFile(tmpFile, []byte(`{"nodes":[],"edges":[]}`), 0644)

	s := newSession()
	s.processLine("load g " + tmpFile)
	if s.active != "g" {
		t.Errorf("expected active 'g', got %q", s.active)
	}
}

func TestProcessLine_Load_MissingArgs(t *testing.T) {
	s := newSession()
	_, _, err := s.processLine("load g")
	if err == nil {
		t.Error("expected error for 'load' with missing file arg")
	}
}

func TestProcessLine_Load_FileNotFound(t *testing.T) {
	s := newSession()
	_, _, err := s.processLine("load g /nonexistent/path.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

// --- save ---

func TestProcessLine_Save_InMemoryNoPath(t *testing.T) {
	s := newSession()
	s.processLine("new g")
	_, _, err := s.processLine("save g")
	if err == nil {
		t.Error("expected error when saving in-memory graph without explicit path")
	}
}

func TestProcessLine_Save_WithExplicitPath(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "out.json")
	s := newSession()
	s.processLine("new g")
	_, _, err := s.processLine(fmt.Sprintf("save g %s", tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("expected save file to exist after save command")
	}
}

func TestProcessLine_Save_UnknownGraph(t *testing.T) {
	s := newSession()
	_, _, err := s.processLine("save nonexistent /tmp/out.json")
	if err == nil {
		t.Error("expected error for saving unknown graph")
	}
}

func TestProcessLine_Save_MissingName(t *testing.T) {
	s := newSession()
	_, _, err := s.processLine("save")
	if err == nil {
		t.Error("expected error for 'save' with no name")
	}
}

// --- DSL queries ---

func TestProcessLine_DSL_NoActiveGraph(t *testing.T) {
	s := newSession()
	_, _, err := s.processLine("MAXPATH FROM A TO B")
	if err == nil {
		t.Error("expected error when no active graph is set")
	}
}

func TestProcessLine_DSL_ReturnsResult(t *testing.T) {
	s := newSession()
	s.processLine("new g")
	s.processLine("CREATE NODE A")
	s.processLine("CREATE NODE B")
	s.processLine("CREATE EDGE e1 FROM A TO B PROB 0.8")

	res, _, err := s.processLine("MAXPATH FROM A TO B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatal("expected a non-nil result")
	}
}

func TestProcessLine_DSL_CorrectProbability(t *testing.T) {
	s := newSession()
	s.processLine("new g")
	s.processLine("CREATE NODE A")
	s.processLine("CREATE NODE B")
	s.processLine("CREATE EDGE e1 FROM A TO B PROB 0.75")

	res, _, err := s.processLine("REACHABILITY FROM A TO B EXACT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pr, ok := res.(probabilistic)
	if !ok {
		t.Fatalf("expected probabilistic result, got %T", res)
	}
	if math.Abs(pr.ProbabilityValue()-0.75) > 0.0001 {
		t.Errorf("expected probability 0.75, got %f", pr.ProbabilityValue())
	}
}

func TestProcessLine_DSL_InvalidSyntax(t *testing.T) {
	s := newSession()
	s.processLine("new g")
	_, _, err := s.processLine("THIS IS NOT VALID DSL")
	if err == nil {
		t.Error("expected error for invalid DSL")
	}
}

func TestProcessLine_DSL_StateAccumulatesAcrossCalls(t *testing.T) {
	// Verify that graph mutations from CREATE persist across processLine calls
	s := newSession()
	s.processLine("new g")
	s.processLine("CREATE NODE A")
	s.processLine("CREATE NODE B")
	s.processLine("CREATE NODE C")
	s.processLine("CREATE EDGE e1 FROM A TO B PROB 0.9")
	s.processLine("CREATE EDGE e2 FROM B TO C PROB 0.8")

	res, _, err := s.processLine("MAXPATH FROM A TO C")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pr, ok := res.(probabilistic)
	if !ok {
		t.Fatalf("expected probabilistic result, got %T", res)
	}
	// A→B→C: 0.9 * 0.8 = 0.72
	if math.Abs(pr.ProbabilityValue()-0.72) > 0.0001 {
		t.Errorf("expected probability 0.72, got %f", pr.ProbabilityValue())
	}
}
