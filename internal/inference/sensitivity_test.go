package inference

import (
	"math"
	"testing"

	"github.com/ritamzico/pgraph/internal/graph"
	"github.com/ritamzico/pgraph/internal/result"
)

func buildSensitivityTestGraph(t *testing.T) graph.ProbabilisticGraphModel {
	t.Helper()
	g := graph.CreateProbAdjListGraph()
	for _, n := range []graph.NodeID{"A", "B", "C", "D"} {
		if err := g.AddNode(n, nil); err != nil {
			t.Fatalf("AddNode %s: %v", n, err)
		}
	}
	edges := []struct {
		id   graph.EdgeID
		from graph.NodeID
		to   graph.NodeID
		prob float64
	}{
		{"eAB", "A", "B", 0.9},
		{"eAC", "A", "C", 0.8},
		{"eBD", "B", "D", 0.7},
		{"eCD", "C", "D", 0.6},
	}
	for _, e := range edges {
		if err := g.AddEdge(e.id, e.from, e.to, e.prob, nil); err != nil {
			t.Fatalf("AddEdge %s: %v", e.id, err)
		}
	}
	return g
}

func TestSensitivityAnalysis_Exact_Baseline(t *testing.T) {
	g := buildSensitivityTestGraph(t)
	res, err := SensitivityAnalysis(g, "A", "D")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path1 := 0.9 * 0.7
	path2 := 0.8 * 0.6
	wantBaseline := 1.0 - (1.0-path1)*(1.0-path2)

	if math.Abs(res.Baseline-wantBaseline) > 1e-9 {
		t.Errorf("baseline: want %.10f, got %.10f", wantBaseline, res.Baseline)
	}
}

func TestSensitivityAnalysis_Exact_ImpactCount(t *testing.T) {
	g := buildSensitivityTestGraph(t)
	res, err := SensitivityAnalysis(g, "A", "D")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Impacts) != 4 {
		t.Errorf("expected 4 impacts (one per edge), got %d", len(res.Impacts))
	}
}

func TestSensitivityAnalysis_Exact_Deltas(t *testing.T) {
	g := buildSensitivityTestGraph(t)
	res, err := SensitivityAnalysis(g, "A", "D")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Build a map for deterministic lookup regardless of sort tie-breaking.
	deltaByEdge := make(map[graph.EdgeID]float64)
	for _, imp := range res.Impacts {
		deltaByEdge[imp.EdgeID] = imp.Delta
	}

	path1 := 0.9 * 0.7 // = 0.63
	path2 := 0.8 * 0.6 // = 0.48
	baseline := 1.0 - (1.0-path1)*(1.0-path2)

	// Removing eAB: only A→C→D survives → reachability = path2
	wantDeltaAB := baseline - path2
	// Removing eAC: only A→B→D survives → reachability = path1
	wantDeltaAC := baseline - path1
	// Removing eBD: only A→C→D survives → reachability = path2
	wantDeltaBD := baseline - path2
	// Removing eCD: only A→B→D survives → reachability = path1
	wantDeltaCD := baseline - path1

	cases := map[graph.EdgeID]float64{
		"eAB": wantDeltaAB,
		"eAC": wantDeltaAC,
		"eBD": wantDeltaBD,
		"eCD": wantDeltaCD,
	}
	for edgeID, wantDelta := range cases {
		got, ok := deltaByEdge[edgeID]
		if !ok {
			t.Errorf("edge %s missing from impacts", edgeID)
			continue
		}
		if math.Abs(got-wantDelta) > 1e-9 {
			t.Errorf("edge %s: want delta %.10f, got %.10f", edgeID, wantDelta, got)
		}
	}
}

func TestSensitivityAnalysis_Exact_SortedDescending(t *testing.T) {
	g := buildSensitivityTestGraph(t)
	res, err := SensitivityAnalysis(g, "A", "D")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 1; i < len(res.Impacts); i++ {
		if res.Impacts[i].Delta > res.Impacts[i-1].Delta {
			t.Errorf("impacts not sorted descending: position %d (%.6f) > position %d (%.6f)",
				i, res.Impacts[i].Delta, i-1, res.Impacts[i-1].Delta)
		}
	}
}

func TestSensitivityAnalysis_Exact_WithoutField(t *testing.T) {
	g := buildSensitivityTestGraph(t)
	res, err := SensitivityAnalysis(g, "A", "D")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, imp := range res.Impacts {
		// Without + Delta should equal Baseline exactly.
		if math.Abs((imp.Without+imp.Delta)-res.Baseline) > 1e-9 {
			t.Errorf("edge %s: Without(%.10f) + Delta(%.10f) != Baseline(%.10f)",
				imp.EdgeID, imp.Without, imp.Delta, res.Baseline)
		}
	}
}

func TestSensitivityAnalysis_Exact_EdgeMetadata(t *testing.T) {
	g := buildSensitivityTestGraph(t)
	res, err := SensitivityAnalysis(g, "A", "D")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	byEdge := make(map[graph.EdgeID]result.EdgeImpact)
	for _, imp := range res.Impacts {
		byEdge[imp.EdgeID] = imp
	}

	checks := []struct {
		id   graph.EdgeID
		from graph.NodeID
		to   graph.NodeID
		prob float64
	}{
		{"eAB", "A", "B", 0.9},
		{"eAC", "A", "C", 0.8},
		{"eBD", "B", "D", 0.7},
		{"eCD", "C", "D", 0.6},
	}
	for _, c := range checks {
		imp, ok := byEdge[c.id]
		if !ok {
			t.Errorf("edge %s missing", c.id)
			continue
		}
		if imp.From != c.from {
			t.Errorf("edge %s: From want %s, got %s", c.id, c.from, imp.From)
		}
		if imp.To != c.to {
			t.Errorf("edge %s: To want %s, got %s", c.id, c.to, imp.To)
		}
		if math.Abs(imp.Probability-c.prob) > 1e-9 {
			t.Errorf("edge %s: Probability want %.3f, got %.3f", c.id, c.prob, imp.Probability)
		}
	}
}

func TestSensitivityAnalysis_Exact_IrrelevantEdgeHasZeroDelta(t *testing.T) {
	// Add an edge X→Y that is disconnected from the A→D query.
	g := buildSensitivityTestGraph(t)
	g.AddNode("X", nil)
	g.AddNode("Y", nil)
	g.AddEdge("eXY", "X", "Y", 0.99, nil)

	res, err := SensitivityAnalysis(g, "A", "D")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, imp := range res.Impacts {
		if imp.EdgeID == "eXY" {
			if math.Abs(imp.Delta) > 1e-9 {
				t.Errorf("irrelevant edge eXY should have delta≈0, got %.10f", imp.Delta)
			}
			return
		}
	}
	t.Error("edge eXY not found in impacts")
}

func TestSensitivityAnalysis_Exact_EmptyGraph(t *testing.T) {
	g := graph.CreateProbAdjListGraph()
	g.AddNode("A", nil)
	g.AddNode("B", nil)
	// No edges.
	res, err := SensitivityAnalysis(g, "A", "B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Baseline != 0.0 {
		t.Errorf("expected baseline 0.0 for disconnected graph, got %f", res.Baseline)
	}
	if len(res.Impacts) != 0 {
		t.Errorf("expected no impacts for empty graph, got %d", len(res.Impacts))
	}
}

func TestSensitivityAnalysis_MonteCarlo_ApproximateBaseline(t *testing.T) {
	g := buildSensitivityTestGraph(t)
	res, err := SensitivityAnalysisMonteCarlo(g, "A", "D", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path1 := 0.9 * 0.7
	path2 := 0.8 * 0.6
	wantBaseline := 1.0 - (1.0-path1)*(1.0-path2)

	// Monte Carlo: allow 3% tolerance.
	if math.Abs(res.Baseline-wantBaseline) > 0.03 {
		t.Errorf("MC baseline: want ≈%.4f, got %.4f", wantBaseline, res.Baseline)
	}
}

func TestSensitivityAnalysis_MonteCarlo_ImpactCount(t *testing.T) {
	g := buildSensitivityTestGraph(t)
	res, err := SensitivityAnalysisMonteCarlo(g, "A", "D", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Impacts) != 4 {
		t.Errorf("expected 4 impacts, got %d", len(res.Impacts))
	}
}

func TestSensitivityAnalysis_MonteCarlo_SortedDescending(t *testing.T) {
	g := buildSensitivityTestGraph(t)
	res, err := SensitivityAnalysisMonteCarlo(g, "A", "D", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 1; i < len(res.Impacts); i++ {
		if res.Impacts[i].Delta > res.Impacts[i-1].Delta {
			t.Errorf("MC impacts not sorted descending at position %d", i)
		}
	}
}
