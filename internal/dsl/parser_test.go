package dsl

import (
	"math"
	"testing"

	"github.com/ritamzico/pgraph/internal/graph"
	"github.com/ritamzico/pgraph/internal/result"
)

func buildTestGraph(t *testing.T) graph.ProbabilisticGraphModel {
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

func TestParser_CreateNode(t *testing.T) {
	baseGraph := graph.CreateProbAdjListGraph()
	parser := CreateParser(baseGraph)

	_, err := parser.ParseLine("CREATE NODE A")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	if !parser.SessionGraph.ContainsNode("A") {
		t.Error("node A should exist after CREATE NODE A")
	}
}

func TestParser_CreateMultipleNodes(t *testing.T) {
	baseGraph := graph.CreateProbAdjListGraph()
	parser := CreateParser(baseGraph)

	_, err := parser.ParseLine("CREATE NODE A, B, C")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	for _, node := range []graph.NodeID{"A", "B", "C"} {
		if !parser.SessionGraph.ContainsNode(node) {
			t.Errorf("node %s should exist", node)
		}
	}
}

func TestParser_CreateEdge(t *testing.T) {
	baseGraph := graph.CreateProbAdjListGraph()
	baseGraph.AddNode("A", nil)
	baseGraph.AddNode("B", nil)

	parser := CreateParser(baseGraph)

	_, err := parser.ParseLine("CREATE EDGE eAB FROM A TO B PROB 0.9")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	if !parser.SessionGraph.ContainsEdgeByID("eAB") {
		t.Error("edge eAB should exist")
	}

	edge, err := parser.SessionGraph.GetEdge("A", "B")
	if err != nil {
		t.Fatalf("GetEdge failed: %v", err)
	}

	if math.Abs(edge.Probability-0.9) > 0.0001 {
		t.Errorf("expected probability 0.9, got %f", edge.Probability)
	}
}

func TestParser_DeleteNode(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	_, err := parser.ParseLine("DELETE NODE A")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	if parser.SessionGraph.ContainsNode("A") {
		t.Error("node A should not exist after DELETE NODE A")
	}
}

func TestParser_DeleteEdge(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	_, err := parser.ParseLine("DELETE EDGE FROM A TO B")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	if parser.SessionGraph.ContainsEdge("A", "B") {
		t.Error("edge A->B should not exist after DELETE EDGE FROM A TO B")
	}
}

func TestParser_DeleteEdgeByID(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	_, err := parser.ParseLine("DELETE EDGE eAB")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	if parser.SessionGraph.ContainsEdgeByID("eAB") {
		t.Error("edge eAB should not exist after DELETE EDGE eAB")
	}
}

func TestParser_MaxPathQuery(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("MAXPATH FROM A TO D")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	pathRes, ok := res.(result.PathResult)
	if !ok {
		t.Fatalf("expected PathResult, got %T", res)
	}

	// Expected: A -> B -> D (0.9 * 0.7 = 0.63)
	expectedProb := 0.9 * 0.7
	if math.Abs(pathRes.Path.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, pathRes.Path.Probability)
	}
}

func TestParser_TopKQuery(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("TOPK FROM A TO D K 2")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	pathsRes, ok := res.(result.PathsResult)
	if !ok {
		t.Fatalf("expected PathsResult, got %T", res)
	}

	if len(pathsRes.Paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(pathsRes.Paths))
	}
}

func TestParser_ReachabilityExact(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("REACHABILITY FROM A TO D EXACT")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// Two paths: A->B->D (0.63) and A->C->D (0.48)
	// Reachability: 1 - (1-0.63)*(1-0.48)
	path1 := 0.9 * 0.7
	path2 := 0.8 * 0.6
	expectedProb := 1.0 - (1.0-path1)*(1.0-path2)

	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}

func TestParser_ReachabilityMonteCarlo(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("REACHABILITY FROM A TO D MONTECARLO")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	sampleRes, ok := res.(result.SampleResult)
	if !ok {
		t.Fatalf("expected SampleResult, got %T", res)
	}

	// Should have an estimate and confidence interval
	if sampleRes.CI95Low > sampleRes.Estimate || sampleRes.Estimate > sampleRes.CI95High {
		t.Errorf("CI bounds invalid: [%f, %f] with estimate %f",
			sampleRes.CI95Low, sampleRes.CI95High, sampleRes.Estimate)
	}
}

func TestParser_ReachabilityDefaultMode(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("REACHABILITY FROM A TO D")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	// Default mode should be EXACT
	_, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult (exact mode), got %T", res)
	}
}

func TestParser_MultiQuery(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("MULTI ( REACHABILITY FROM A TO B EXACT, REACHABILITY FROM A TO C EXACT )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	multiRes, ok := res.(result.MultiResult)
	if !ok {
		t.Fatalf("expected MultiResult, got %T", res)
	}

	if len(multiRes.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(multiRes.Results))
	}
}

func TestParser_AndQuery(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("AND ( REACHABILITY FROM A TO B EXACT, REACHABILITY FROM A TO C EXACT )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// AND(0.9, 0.8) = 0.72
	expectedProb := 0.9 * 0.8
	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}

func TestParser_OrQuery(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("OR ( REACHABILITY FROM A TO B EXACT, REACHABILITY FROM A TO C EXACT )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// OR(0.9, 0.8) = 1 - (1-0.9)*(1-0.8) = 0.98
	expectedProb := 1.0 - (1.0-0.9)*(1.0-0.8)
	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}

func TestParser_ThresholdQuery(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("THRESHOLD 0.85 ( REACHABILITY FROM A TO B EXACT )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	boolRes, ok := res.(result.BooleanResult)
	if !ok {
		t.Fatalf("expected BooleanResult, got %T", res)
	}

	// 0.9 >= 0.85, should be true
	if !boolRes.Value {
		t.Error("expected true (0.9 >= 0.85), got false")
	}
}

func TestParser_ThresholdQueryFalse(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("THRESHOLD 0.95 ( REACHABILITY FROM A TO B EXACT )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	boolRes, ok := res.(result.BooleanResult)
	if !ok {
		t.Fatalf("expected BooleanResult, got %T", res)
	}

	// 0.9 < 0.95, should be false
	if boolRes.Value {
		t.Error("expected false (0.9 < 0.95), got true")
	}
}

func TestParser_ConditionalQueryInactiveEdge(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("CONDITIONAL GIVEN EDGE eAB INACTIVE ( REACHABILITY FROM A TO D EXACT )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// With edge A->B inactive, only path is A->C->D (0.8 * 0.6 = 0.48)
	expectedProb := 0.8 * 0.6
	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}

func TestParser_ConditionalQueryInactiveNode(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("CONDITIONAL GIVEN NODE B INACTIVE ( REACHABILITY FROM A TO D EXACT )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// With node B inactive, only path is A->C->D (0.8 * 0.6 = 0.48)
	expectedProb := 0.8 * 0.6
	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}

func TestParser_ConditionalQueryMultipleConditions(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("CONDITIONAL GIVEN EDGE eAB INACTIVE, EDGE eCD INACTIVE ( REACHABILITY FROM A TO D EXACT )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// With A->B and C->D inactive, only path is via... actually there's only B->D left
	// So there's no complete path from A to D
	if probRes.Probability != 0.0 {
		t.Errorf("expected probability 0.0 (no path), got %f", probRes.Probability)
	}
}

func TestParser_NestedCompositeQueries(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	// This is tricky - the DSL doesn't support direct nesting in one line easily
	// But we can test multi-level structures
	res, err := parser.ParseLine("AND ( REACHABILITY FROM A TO B EXACT, REACHABILITY FROM C TO D EXACT )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// AND(0.9, 0.6) = 0.54
	expectedProb := 0.9 * 0.6
	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}

func TestParser_CaseInsensitivity(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	testCases := []string{
		"maxpath from A to D",
		"MAXPATH FROM A TO D",
		"MaxPath From A To D",
		"MaXpAtH fRoM A tO D",
	}

	for _, tc := range testCases {
		res, err := parser.ParseLine(tc)
		if err != nil {
			t.Errorf("ParseLine failed for %q: %v", tc, err)
			continue
		}

		if _, ok := res.(result.PathResult); !ok {
			t.Errorf("expected PathResult for %q, got %T", tc, res)
		}
	}
}

func TestParser_InvalidSyntax(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	testCases := []string{
		"MAXPATH A D",               // Missing FROM/TO
		"CREATE NODE",               // Missing node IDs
		"REACHABILITY FROM A",       // Missing TO
		"TOPK FROM A TO B",          // Missing K
		"THRESHOLD ( MAXPATH A D )", // Missing threshold value
		"AND ( )",                   // Empty query list
		"FOOBAR",                    // Unknown command
	}

	for _, tc := range testCases {
		_, err := parser.ParseLine(tc)
		if err == nil {
			t.Errorf("expected error for invalid syntax %q, got nil", tc)
		}
	}
}

func TestParser_NonexistentNode(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	// Try to create edge with nonexistent node
	_, err := parser.ParseLine("CREATE EDGE eXY FROM X TO Y PROB 0.5")
	if err == nil {
		t.Error("expected error when creating edge with nonexistent nodes")
	}
}

func TestParser_DuplicateNode(t *testing.T) {
	baseGraph := graph.CreateProbAdjListGraph()
	parser := CreateParser(baseGraph)

	_, err := parser.ParseLine("CREATE NODE A")
	if err != nil {
		t.Fatalf("first CREATE NODE A failed: %v", err)
	}

	_, err = parser.ParseLine("CREATE NODE A")
	if err == nil {
		t.Error("expected error when creating duplicate node")
	}
}

func TestParser_ComplexSupplyChainScenario(t *testing.T) {
	baseGraph := graph.CreateProbAdjListGraph()
	parser := CreateParser(baseGraph)

	// Build supply chain incrementally via DSL
	commands := []string{
		"CREATE NODE Mine, Factory, Warehouse, Store",
		"CREATE EDGE e1 FROM Mine TO Factory PROB 0.95",
		"CREATE EDGE e2 FROM Factory TO Warehouse PROB 0.90",
		"CREATE EDGE e3 FROM Warehouse TO Store PROB 0.88",
	}

	for _, cmd := range commands {
		if _, err := parser.ParseLine(cmd); err != nil {
			t.Fatalf("command %q failed: %v", cmd, err)
		}
	}

	// Query end-to-end reachability
	res, err := parser.ParseLine("REACHABILITY FROM Mine TO Store EXACT")
	if err != nil {
		t.Fatalf("reachability query failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	expectedProb := 0.95 * 0.90 * 0.88
	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}

func TestParser_ConditionalWithThreshold(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	// Test nested: THRESHOLD over CONDITIONAL
	res, err := parser.ParseLine("THRESHOLD 0.5 ( CONDITIONAL GIVEN EDGE eAB INACTIVE ( REACHABILITY FROM A TO D EXACT ) )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	boolRes, ok := res.(result.BooleanResult)
	if !ok {
		t.Fatalf("expected BooleanResult, got %T", res)
	}

	// With eAB inactive, reachability A->D is 0.48 (< 0.5), so should be false
	if boolRes.Value {
		t.Error("expected false (0.48 < 0.5), got true")
	}
}

// --- AGGREGATE query tests ---

func TestParser_AggregateMean(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("AGGREGATE MEAN ( REACHABILITY FROM A TO B EXACT, REACHABILITY FROM A TO C EXACT )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// Mean of 0.9 and 0.8 = 0.85
	if math.Abs(probRes.Probability-0.85) > 0.0001 {
		t.Errorf("expected 0.85, got %f", probRes.Probability)
	}
}

func TestParser_AggregateMax(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("AGGREGATE MAX ( REACHABILITY FROM A TO B EXACT, REACHABILITY FROM A TO C EXACT, REACHABILITY FROM B TO D EXACT )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// Max of 0.9, 0.8, 0.7 = 0.9
	if math.Abs(probRes.Probability-0.9) > 0.0001 {
		t.Errorf("expected 0.9, got %f", probRes.Probability)
	}
}

func TestParser_AggregateMin(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("AGGREGATE MIN ( REACHABILITY FROM A TO B EXACT, REACHABILITY FROM A TO C EXACT, REACHABILITY FROM B TO D EXACT )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// Min of 0.9, 0.8, 0.7 = 0.7
	if math.Abs(probRes.Probability-0.7) > 0.0001 {
		t.Errorf("expected 0.7, got %f", probRes.Probability)
	}
}

func TestParser_AggregateBestPath(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("AGGREGATE BESTPATH ( MAXPATH FROM A TO D, MAXPATH FROM A TO B )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	pathRes, ok := res.(result.PathResult)
	if !ok {
		t.Fatalf("expected PathResult, got %T", res)
	}

	// A→D best path = A→B→D (0.63), A→B = 0.9; best is 0.9
	if math.Abs(pathRes.Path.Probability-0.9) > 0.0001 {
		t.Errorf("expected 0.9, got %f", pathRes.Path.Probability)
	}
}

func TestParser_AggregateCountAbove(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("AGGREGATE COUNTABOVE 0.75 ( REACHABILITY FROM A TO B EXACT, REACHABILITY FROM A TO C EXACT, REACHABILITY FROM B TO D EXACT )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// 0.9 >= 0.75 ✓, 0.8 >= 0.75 ✓, 0.7 >= 0.75 ✗ → 2/3
	expected := 2.0 / 3.0
	if math.Abs(probRes.Probability-expected) > 0.0001 {
		t.Errorf("expected %f, got %f", expected, probRes.Probability)
	}
}

func TestParser_AggregateCaseInsensitivity(t *testing.T) {
	baseGraph := buildTestGraph(t)

	cases := []string{
		"aggregate mean ( REACHABILITY FROM A TO B EXACT, REACHABILITY FROM A TO C EXACT )",
		"AGGREGATE MEAN ( reachability from A to B exact, reachability from A to C exact )",
		"Aggregate Mean ( Reachability From A To B Exact, Reachability From A To C Exact )",
	}

	for _, tc := range cases {
		parser := CreateParser(baseGraph)
		res, err := parser.ParseLine(tc)
		if err != nil {
			t.Errorf("ParseLine failed for %q: %v", tc, err)
			continue
		}

		probRes, ok := res.(result.ProbabilityResult)
		if !ok {
			t.Errorf("expected ProbabilityResult for %q, got %T", tc, res)
			continue
		}

		if math.Abs(probRes.Probability-0.85) > 0.0001 {
			t.Errorf("expected 0.85 for %q, got %f", tc, probRes.Probability)
		}
	}
}

func TestParser_AggregateNestedInThreshold(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	// THRESHOLD over AGGREGATE: mean of 0.9, 0.8 = 0.85 >= 0.8 → true
	res, err := parser.ParseLine("THRESHOLD 0.8 ( AGGREGATE MEAN ( REACHABILITY FROM A TO B EXACT, REACHABILITY FROM A TO C EXACT ) )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	boolRes, ok := res.(result.BooleanResult)
	if !ok {
		t.Fatalf("expected BooleanResult, got %T", res)
	}

	if !boolRes.Value {
		t.Error("expected true (0.85 >= 0.8), got false")
	}
}

func TestParser_AggregateWithConditionalSubquery(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	// Aggregate with a conditional inner query
	res, err := parser.ParseLine("AGGREGATE MAX ( REACHABILITY FROM A TO D EXACT, CONDITIONAL GIVEN EDGE eAB INACTIVE ( REACHABILITY FROM A TO D EXACT ) )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// Full graph reachability A→D ≈ 0.8076, with eAB inactive = 0.48
	// MAX(0.8076, 0.48) = 0.8076
	path1 := 0.9 * 0.7
	path2 := 0.8 * 0.6
	fullReachability := 1.0 - (1.0-path1)*(1.0-path2)

	if math.Abs(probRes.Probability-fullReachability) > 0.0001 {
		t.Errorf("expected %f, got %f", fullReachability, probRes.Probability)
	}
}

func TestParser_AggregateSingleQuery(t *testing.T) {
	baseGraph := buildTestGraph(t)
	parser := CreateParser(baseGraph)

	res, err := parser.ParseLine("AGGREGATE MIN ( REACHABILITY FROM A TO B EXACT )")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	if math.Abs(probRes.Probability-0.9) > 0.0001 {
		t.Errorf("expected 0.9, got %f", probRes.Probability)
	}
}

// ── Property tests ──────────────────────────────────────────────────────

func TestParser_CreateNodeWithProperties(t *testing.T) {
	baseGraph := graph.CreateProbAdjListGraph()
	parser := CreateParser(baseGraph)

	_, err := parser.ParseLine(`CREATE NODE supplier { region: "US", risk_score: 0.85, count: 42, is_active: true }`)
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	if !parser.SessionGraph.ContainsNode("supplier") {
		t.Fatal("node supplier should exist")
	}

	nodes := parser.SessionGraph.GetNodes()
	var node *graph.Node
	for _, n := range nodes {
		if n.ID == "supplier" {
			node = n
			break
		}
	}
	if node == nil {
		t.Fatal("could not find node supplier")
	}

	// String property
	if v, ok := node.Props["region"]; !ok {
		t.Error("missing property region")
	} else if v.Kind != graph.StringVal || v.S != "US" {
		t.Errorf("expected StringVal US, got %+v", v)
	}

	// Float property
	if v, ok := node.Props["risk_score"]; !ok {
		t.Error("missing property risk_score")
	} else if v.Kind != graph.FloatVal || math.Abs(v.F-0.85) > 0.0001 {
		t.Errorf("expected FloatVal 0.85, got %+v", v)
	}

	// Int property
	if v, ok := node.Props["count"]; !ok {
		t.Error("missing property count")
	} else if v.Kind != graph.IntVal || v.I != 42 {
		t.Errorf("expected IntVal 42, got %+v", v)
	}

	// Bool property (true)
	if v, ok := node.Props["is_active"]; !ok {
		t.Error("missing property is_active")
	} else if v.Kind != graph.BoolVal || !v.B {
		t.Errorf("expected BoolVal true, got %+v", v)
	}
}

func TestParser_CreateNodeWithBoolFalse(t *testing.T) {
	baseGraph := graph.CreateProbAdjListGraph()
	parser := CreateParser(baseGraph)

	_, err := parser.ParseLine(`CREATE NODE x { enabled: false }`)
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	nodes := parser.SessionGraph.GetNodes()
	var node *graph.Node
	for _, n := range nodes {
		if n.ID == "x" {
			node = n
			break
		}
	}
	if node == nil {
		t.Fatal("could not find node x")
	}

	if v, ok := node.Props["enabled"]; !ok {
		t.Error("missing property enabled")
	} else if v.Kind != graph.BoolVal || v.B {
		t.Errorf("expected BoolVal false, got %+v", v)
	}
}

func TestParser_CreateNodeWithoutProperties(t *testing.T) {
	baseGraph := graph.CreateProbAdjListGraph()
	parser := CreateParser(baseGraph)

	_, err := parser.ParseLine("CREATE NODE A")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	if !parser.SessionGraph.ContainsNode("A") {
		t.Error("node A should exist")
	}

	nodes := parser.SessionGraph.GetNodes()
	for _, n := range nodes {
		if n.ID == "A" && n.Props != nil {
			t.Errorf("expected nil props for node without properties, got %v", n.Props)
		}
	}
}

func TestParser_CreateMultipleNodesWithProperties(t *testing.T) {
	baseGraph := graph.CreateProbAdjListGraph()
	parser := CreateParser(baseGraph)

	_, err := parser.ParseLine(`CREATE NODE a, b, c { type: "warehouse" }`)
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	for _, id := range []graph.NodeID{"a", "b", "c"} {
		if !parser.SessionGraph.ContainsNode(id) {
			t.Errorf("node %s should exist", id)
		}
	}

	nodes := parser.SessionGraph.GetNodes()
	for _, n := range nodes {
		v, ok := n.Props["type"]
		if !ok {
			t.Errorf("node %s missing property type", n.ID)
		} else if v.Kind != graph.StringVal || v.S != "warehouse" {
			t.Errorf("node %s: expected StringVal warehouse, got %+v", n.ID, v)
		}
	}
}

func TestParser_CreateEdgeWithProperties(t *testing.T) {
	baseGraph := graph.CreateProbAdjListGraph()
	baseGraph.AddNode("A", nil)
	baseGraph.AddNode("B", nil)
	parser := CreateParser(baseGraph)

	_, err := parser.ParseLine(`CREATE EDGE eAB FROM A TO B PROB 0.9 { distance: 100, transport: "truck" }`)
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	edge, err := parser.SessionGraph.GetEdge("A", "B")
	if err != nil {
		t.Fatalf("GetEdge failed: %v", err)
	}

	// Int property
	if v, ok := edge.Props["distance"]; !ok {
		t.Error("missing property distance")
	} else if v.Kind != graph.IntVal || v.I != 100 {
		t.Errorf("expected IntVal 100, got %+v", v)
	}

	// String property
	if v, ok := edge.Props["transport"]; !ok {
		t.Error("missing property transport")
	} else if v.Kind != graph.StringVal || v.S != "truck" {
		t.Errorf("expected StringVal truck, got %+v", v)
	}
}

func TestParser_CreateEdgeWithoutProperties(t *testing.T) {
	baseGraph := graph.CreateProbAdjListGraph()
	baseGraph.AddNode("A", nil)
	baseGraph.AddNode("B", nil)
	parser := CreateParser(baseGraph)

	_, err := parser.ParseLine("CREATE EDGE eAB FROM A TO B PROB 0.9")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	edge, err := parser.SessionGraph.GetEdge("A", "B")
	if err != nil {
		t.Fatalf("GetEdge failed: %v", err)
	}

	if edge.Props != nil {
		t.Errorf("expected nil props for edge without properties, got %v", edge.Props)
	}
}

func TestParser_PropertyKeywordsCaseInsensitive(t *testing.T) {
	baseGraph := graph.CreateProbAdjListGraph()
	parser := CreateParser(baseGraph)

	_, err := parser.ParseLine(`CREATE NODE n { flag: TRUE, other: FALSE }`)
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	nodes := parser.SessionGraph.GetNodes()
	var node *graph.Node
	for _, n := range nodes {
		if n.ID == "n" {
			node = n
			break
		}
	}
	if node == nil {
		t.Fatal("could not find node n")
	}

	if v := node.Props["flag"]; v.Kind != graph.BoolVal || !v.B {
		t.Errorf("expected BoolVal true, got %+v", v)
	}
	if v := node.Props["other"]; v.Kind != graph.BoolVal || v.B {
		t.Errorf("expected BoolVal false, got %+v", v)
	}
}

// ── Case sensitivity and identifier tests ───────────────────────────────

func TestParser_NodeNamesCaseSensitive(t *testing.T) {
	baseGraph := graph.CreateProbAdjListGraph()
	parser := CreateParser(baseGraph)

	// Create two nodes whose names differ only in case
	_, err := parser.ParseLine("CREATE NODE NodeA")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}
	_, err = parser.ParseLine("CREATE NODE nodea")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	if !parser.SessionGraph.ContainsNode("NodeA") {
		t.Error("node NodeA should exist")
	}
	if !parser.SessionGraph.ContainsNode("nodea") {
		t.Error("node nodea should exist")
	}

	// They must be distinct nodes
	nodes := parser.SessionGraph.GetNodes()
	if len(nodes) != 2 {
		t.Errorf("expected 2 distinct nodes, got %d", len(nodes))
	}
}

func TestParser_EdgeNamesCaseSensitive(t *testing.T) {
	baseGraph := graph.CreateProbAdjListGraph()
	baseGraph.AddNode("A", nil)
	baseGraph.AddNode("B", nil)
	baseGraph.AddNode("C", nil)
	parser := CreateParser(baseGraph)

	_, err := parser.ParseLine("CREATE EDGE MyEdge FROM A TO B PROB 0.9")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}
	_, err = parser.ParseLine("CREATE EDGE myedge FROM A TO C PROB 0.8")
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	edgeAB, err := parser.SessionGraph.GetEdge("A", "B")
	if err != nil {
		t.Fatalf("GetEdge A->B failed: %v", err)
	}
	edgeAC, err := parser.SessionGraph.GetEdge("A", "C")
	if err != nil {
		t.Fatalf("GetEdge A->C failed: %v", err)
	}

	if edgeAB.ID != "MyEdge" {
		t.Errorf("expected edge ID MyEdge, got %s", edgeAB.ID)
	}
	if edgeAC.ID != "myedge" {
		t.Errorf("expected edge ID myedge, got %s", edgeAC.ID)
	}
}

func TestParser_KeywordsCaseInsensitiveInStatements(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"lowercase create", "create node X"},
		{"uppercase CREATE", "CREATE NODE X"},
		{"mixed case CrEaTe", "CrEaTe NoDe X"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseGraph := graph.CreateProbAdjListGraph()
			parser := CreateParser(baseGraph)

			_, err := parser.ParseLine(tc.input)
			if err != nil {
				t.Fatalf("ParseLine failed for %q: %v", tc.input, err)
			}

			if !parser.SessionGraph.ContainsNode("X") {
				t.Errorf("node X should exist after %q", tc.input)
			}
		})
	}
}

func TestParser_KeywordsCaseInsensitiveInDelete(t *testing.T) {
	testCases := []string{
		"delete node A",
		"DELETE NODE A",
		"DeLeTe NoDe A",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			baseGraph := graph.CreateProbAdjListGraph()
			baseGraph.AddNode("A", nil)
			parser := CreateParser(baseGraph)

			_, err := parser.ParseLine(tc)
			if err != nil {
				t.Fatalf("ParseLine failed for %q: %v", tc, err)
			}

			if parser.SessionGraph.ContainsNode("A") {
				t.Errorf("node A should be deleted after %q", tc)
			}
		})
	}
}

func TestParser_InvalidCharactersInNodeName(t *testing.T) {
	invalidNames := []string{
		"CREATE NODE node-name",   // hyphen
		"CREATE NODE node.name",   // dot
		"CREATE NODE node@name",   // at sign
		"CREATE NODE node name",   // space (parses as two separate idents)
		"CREATE NODE 123abc",      // starts with digit
		"CREATE NODE node!",       // exclamation
	}

	for _, tc := range invalidNames {
		t.Run(tc, func(t *testing.T) {
			baseGraph := graph.CreateProbAdjListGraph()
			parser := CreateParser(baseGraph)

			_, err := parser.ParseLine(tc)
			if err == nil {
				t.Errorf("expected error for invalid identifier in %q, got nil", tc)
			}
		})
	}
}

func TestParser_ValidIdentifierPatterns(t *testing.T) {
	validNames := []struct {
		name  string
		input string
	}{
		{"lowercase", "CREATE NODE abc"},
		{"uppercase", "CREATE NODE ABC"},
		{"mixed case", "CREATE NODE AbC"},
		{"with underscore", "CREATE NODE my_node"},
		{"leading underscore", "CREATE NODE _private"},
		{"with digits", "CREATE NODE node42"},
		{"underscore and digits", "CREATE NODE _n0d3"},
		{"single letter", "CREATE NODE x"},
	}

	for _, tc := range validNames {
		t.Run(tc.name, func(t *testing.T) {
			baseGraph := graph.CreateProbAdjListGraph()
			parser := CreateParser(baseGraph)

			_, err := parser.ParseLine(tc.input)
			if err != nil {
				t.Fatalf("ParseLine failed for %q: %v", tc.input, err)
			}
		})
	}
}

func TestParser_KeywordAsNodeNameRejected(t *testing.T) {
	// Keywords cannot be used as node/edge names because the lexer
	// classifies them as Keyword tokens, not Ident tokens.
	keywords := []string{
		"CREATE NODE create",
		"CREATE NODE delete",
		"CREATE NODE from",
		"CREATE NODE edge",
		"CREATE NODE true",
		"CREATE NODE false",
		"CREATE NODE maxpath",
		"CREATE NODE reachability",
	}

	for _, tc := range keywords {
		t.Run(tc, func(t *testing.T) {
			baseGraph := graph.CreateProbAdjListGraph()
			parser := CreateParser(baseGraph)

			_, err := parser.ParseLine(tc)
			if err == nil {
				t.Errorf("expected error when using keyword as node name in %q, got nil", tc)
			}
		})
	}
}
