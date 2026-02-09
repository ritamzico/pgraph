package query

import (
	"testing"

	"github.com/ritamzico/pgraph/internal/graph"
)

// buildLinearGraph creates A -> B -> C with specified probabilities
func buildLinearGraph(t *testing.T, pAB, pBC float64) graph.ProbabilisticGraphModel {
	t.Helper()
	g := graph.CreateProbAdjListGraph()

	if err := g.AddNode("A", nil); err != nil {
		t.Fatalf("failed to add node A: %v", err)
	}
	if err := g.AddNode("B", nil); err != nil {
		t.Fatalf("failed to add node B: %v", err)
	}
	if err := g.AddNode("C", nil); err != nil {
		t.Fatalf("failed to add node C: %v", err)
	}

	if err := g.AddEdge("eAB", "A", "B", pAB, nil); err != nil {
		t.Fatalf("failed to add edge A->B: %v", err)
	}
	if err := g.AddEdge("eBC", "B", "C", pBC, nil); err != nil {
		t.Fatalf("failed to add edge B->C: %v", err)
	}

	return g
}

// buildDiamondGraph creates A -> B,C -> D with multiple paths
func buildDiamondGraph(t *testing.T) graph.ProbabilisticGraphModel {
	t.Helper()
	g := graph.CreateProbAdjListGraph()

	nodes := []graph.NodeID{"A", "B", "C", "D"}
	for _, n := range nodes {
		if err := g.AddNode(n, nil); err != nil {
			t.Fatalf("failed to add node %s: %v", n, err)
		}
	}

	edges := []struct {
		id   string
		from string
		to   string
		prob float64
	}{
		{"eAB", "A", "B", 0.9},
		{"eAC", "A", "C", 0.8},
		{"eBD", "B", "D", 0.7},
		{"eCD", "C", "D", 0.6},
	}

	for _, e := range edges {
		if err := g.AddEdge(graph.EdgeID(e.id), graph.NodeID(e.from), graph.NodeID(e.to), e.prob, nil); err != nil {
			t.Fatalf("failed to add edge %s->%s: %v", e.from, e.to, err)
		}
	}

	return g
}

// buildSupplyChainGraph creates a realistic supply chain scenario
func buildSupplyChainGraph(t *testing.T) graph.ProbabilisticGraphModel {
	t.Helper()
	g := graph.CreateProbAdjListGraph()

	nodes := []graph.NodeID{
		"LithiumMine", "BatteryFactory", "CarFactory", "Dealer",
		"BackupSupplier", "AlternativeDealer",
	}
	for _, n := range nodes {
		if err := g.AddNode(n, nil); err != nil {
			t.Fatalf("failed to add node %s: %v", n, err)
		}
	}

	edges := []struct {
		id   string
		from string
		to   string
		prob float64
	}{
		{"e1", "LithiumMine", "BatteryFactory", 0.95},
		{"e2", "BackupSupplier", "BatteryFactory", 0.85},
		{"e3", "BatteryFactory", "CarFactory", 0.90},
		{"e4", "CarFactory", "Dealer", 0.88},
		{"e5", "CarFactory", "AlternativeDealer", 0.92},
	}

	for _, e := range edges {
		if err := g.AddEdge(graph.EdgeID(e.id), graph.NodeID(e.from), graph.NodeID(e.to), e.prob, nil); err != nil {
			t.Fatalf("failed to add edge %s->%s: %v", e.from, e.to, err)
		}
	}

	return g
}

// buildDisconnectedGraph creates two separate components
func buildDisconnectedGraph(t *testing.T) graph.ProbabilisticGraphModel {
	t.Helper()
	g := graph.CreateProbAdjListGraph()

	// Component 1: A -> B
	if err := g.AddNode("A", nil); err != nil {
		t.Fatalf("failed to add node A: %v", err)
	}
	if err := g.AddNode("B", nil); err != nil {
		t.Fatalf("failed to add node B: %v", err)
	}
	if err := g.AddEdge("eAB", "A", "B", 0.9, nil); err != nil {
		t.Fatalf("failed to add edge A->B: %v", err)
	}

	// Component 2: X -> Y
	if err := g.AddNode("X", nil); err != nil {
		t.Fatalf("failed to add node X: %v", err)
	}
	if err := g.AddNode("Y", nil); err != nil {
		t.Fatalf("failed to add node Y: %v", err)
	}
	if err := g.AddEdge("eXY", "X", "Y", 0.8, nil); err != nil {
		t.Fatalf("failed to add edge X->Y: %v", err)
	}

	return g
}

// buildSingleNodeGraph creates a graph with just one node
func buildSingleNodeGraph(t *testing.T) graph.ProbabilisticGraphModel {
	t.Helper()
	g := graph.CreateProbAdjListGraph()

	if err := g.AddNode("A", nil); err != nil {
		t.Fatalf("failed to add node A: %v", err)
	}

	return g
}

// buildComplexGraph creates a graph with multiple parallel paths and cycles
func buildComplexGraph(t *testing.T) graph.ProbabilisticGraphModel {
	t.Helper()
	g := graph.CreateProbAdjListGraph()

	nodes := []graph.NodeID{"A", "B", "C", "D", "E", "F"}
	for _, n := range nodes {
		if err := g.AddNode(n, nil); err != nil {
			t.Fatalf("failed to add node %s: %v", n, err)
		}
	}

	edges := []struct {
		id   string
		from string
		to   string
		prob float64
	}{
		{"eAB", "A", "B", 0.9},
		{"eAC", "A", "C", 0.85},
		{"eAD", "A", "D", 0.7},
		{"eBE", "B", "E", 0.8},
		{"eCE", "C", "E", 0.75},
		{"eDE", "D", "E", 0.65},
		{"eEF", "E", "F", 0.95},
		{"eBF", "B", "F", 0.6},
	}

	for _, e := range edges {
		if err := g.AddEdge(graph.EdgeID(e.id), graph.NodeID(e.from), graph.NodeID(e.to), e.prob, nil); err != nil {
			t.Fatalf("failed to add edge %s->%s: %v", e.from, e.to, err)
		}
	}

	return g
}
