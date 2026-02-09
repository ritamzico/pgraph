package query

import (
	"context"
	"math"
	"testing"

	"github.com/ritamzico/pgraph/internal/result"
)

func TestMaxProbabilityPathQuery_LinearGraph(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)
	q := MaxProbabilityPathQuery{Start: "A", End: "C"}

	res, err := q.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	pathRes, ok := res.(result.PathResult)
	if !ok {
		t.Fatalf("expected PathResult, got %T", res)
	}

	// Expected path: A -> B -> C with probability 0.9 * 0.8 = 0.72
	expectedProb := 0.9 * 0.8
	if math.Abs(pathRes.Path.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, pathRes.Path.Probability)
	}

	if len(pathRes.Path.NodeIDs) != 3 {
		t.Errorf("expected path length 3, got %d", len(pathRes.Path.NodeIDs))
	}
}

func TestMaxProbabilityPathQuery_DiamondGraph(t *testing.T) {
	g := buildDiamondGraph(t)
	q := MaxProbabilityPathQuery{Start: "A", End: "D"}

	res, err := q.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	pathRes, ok := res.(result.PathResult)
	if !ok {
		t.Fatalf("expected PathResult, got %T", res)
	}

	// Expected path: A -> B -> D with probability 0.9 * 0.7 = 0.63
	// (higher than A -> C -> D which is 0.8 * 0.6 = 0.48)
	expectedProb := 0.9 * 0.7
	if math.Abs(pathRes.Path.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, pathRes.Path.Probability)
	}
}

func TestMaxProbabilityPathQuery_NoPath(t *testing.T) {
	g := buildDisconnectedGraph(t)
	q := MaxProbabilityPathQuery{Start: "A", End: "X"}

	res, err := q.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	pathRes, ok := res.(result.PathResult)
	if !ok {
		t.Fatalf("expected PathResult, got %T", res)
	}

	if len(pathRes.Path.NodeIDs) != 0 {
		t.Errorf("expected empty path for disconnected nodes, got %+v", pathRes.Path)
	}
}

func TestMaxProbabilityPathQuery_SameNode(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)
	q := MaxProbabilityPathQuery{Start: "A", End: "A"}

	res, err := q.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	pathRes, ok := res.(result.PathResult)
	if !ok {
		t.Fatalf("expected PathResult, got %T", res)
	}

	// Path from A to A should have probability 1.0 and length 1
	if pathRes.Path.Probability != 1.0 {
		t.Errorf("expected probability 1.0 for same node, got %f", pathRes.Path.Probability)
	}
	if len(pathRes.Path.NodeIDs) != 1 {
		t.Errorf("expected path length 1, got %d", len(pathRes.Path.NodeIDs))
	}
}

func TestTopKProbabilityPathsQuery_DiamondGraph(t *testing.T) {
	g := buildDiamondGraph(t)
	q := TopKProbabilityPathsQuery{Start: "A", End: "D", K: 2}

	res, err := q.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	pathsRes, ok := res.(result.PathsResult)
	if !ok {
		t.Fatalf("expected PathsResult, got %T", res)
	}

	if len(pathsRes.Paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(pathsRes.Paths))
	}

	// First path should be A -> B -> D (0.63)
	if math.Abs(pathsRes.Paths[0].Probability-0.63) > 0.0001 {
		t.Errorf("expected first path probability 0.63, got %f", pathsRes.Paths[0].Probability)
	}

	// Second path should be A -> C -> D (0.48)
	if math.Abs(pathsRes.Paths[1].Probability-0.48) > 0.0001 {
		t.Errorf("expected second path probability 0.48, got %f", pathsRes.Paths[1].Probability)
	}

	// Verify paths are in descending order
	if pathsRes.Paths[0].Probability < pathsRes.Paths[1].Probability {
		t.Error("paths should be sorted in descending order by probability")
	}
}

func TestTopKProbabilityPathsQuery_KLargerThanAvailable(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)
	q := TopKProbabilityPathsQuery{Start: "A", End: "C", K: 10}

	res, err := q.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	pathsRes, ok := res.(result.PathsResult)
	if !ok {
		t.Fatalf("expected PathsResult, got %T", res)
	}

	// Only one path exists A -> B -> C
	if len(pathsRes.Paths) != 1 {
		t.Errorf("expected 1 path (only path available), got %d", len(pathsRes.Paths))
	}
}

func TestTopKProbabilityPathsQuery_ComplexGraph(t *testing.T) {
	g := buildComplexGraph(t)
	q := TopKProbabilityPathsQuery{Start: "A", End: "F", K: 5}

	res, err := q.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	pathsRes, ok := res.(result.PathsResult)
	if !ok {
		t.Fatalf("expected PathsResult, got %T", res)
	}

	if len(pathsRes.Paths) == 0 {
		t.Fatal("expected at least one path")
	}

	// Verify all paths start at A and end at F
	for i, p := range pathsRes.Paths {
		if p.NodeIDs[0] != "A" {
			t.Errorf("path %d should start at A, got %s", i, p.NodeIDs[0])
		}
		if p.NodeIDs[len(p.NodeIDs)-1] != "F" {
			t.Errorf("path %d should end at F, got %s", i, p.NodeIDs[len(p.NodeIDs)-1])
		}
	}

	// Verify paths are sorted
	for i := 0; i < len(pathsRes.Paths)-1; i++ {
		if pathsRes.Paths[i].Probability < pathsRes.Paths[i+1].Probability {
			t.Errorf("paths not sorted: path[%d]=%.4f > path[%d]=%.4f",
				i, pathsRes.Paths[i].Probability, i+1, pathsRes.Paths[i+1].Probability)
		}
	}
}

func TestReachabilityProbabilityQuery_Exact_LinearGraph(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)
	q := ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact}

	res, err := q.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// Reachability probability: A -> B -> C = 0.9 * 0.8 = 0.72
	expectedProb := 0.9 * 0.8
	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}

func TestReachabilityProbabilityQuery_Exact_DiamondGraph(t *testing.T) {
	g := buildDiamondGraph(t)
	q := ReachabilityProbabilityQuery{Start: "A", End: "D", Mode: Exact}

	res, err := q.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// Two paths: A->B->D (0.9*0.7=0.63) and A->C->D (0.8*0.6=0.48)
	// Reachability = 1 - (1-0.63)*(1-0.48) = 1 - 0.37*0.52 = 1 - 0.1924 = 0.8076
	path1 := 0.9 * 0.7
	path2 := 0.8 * 0.6
	expectedProb := 1.0 - (1.0-path1)*(1.0-path2)

	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}

func TestReachabilityProbabilityQuery_Exact_NoPath(t *testing.T) {
	g := buildDisconnectedGraph(t)
	q := ReachabilityProbabilityQuery{Start: "A", End: "X", Mode: Exact}

	res, err := q.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	if probRes.Probability != 0.0 {
		t.Errorf("expected probability 0.0 for no path, got %f", probRes.Probability)
	}
}

func TestReachabilityProbabilityQuery_Exact_SameNode(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)
	q := ReachabilityProbabilityQuery{Start: "A", End: "A", Mode: Exact}

	res, err := q.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	if probRes.Probability != 1.0 {
		t.Errorf("expected probability 1.0 for same node, got %f", probRes.Probability)
	}
}

func TestReachabilityProbabilityQuery_MonteCarlo_LinearGraph(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)
	q := ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: MonteCarlo}

	res, err := q.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	sampleRes, ok := res.(result.SampleResult)
	if !ok {
		t.Fatalf("expected SampleResult, got %T", res)
	}

	// Expected probability: 0.9 * 0.8 = 0.72
	// Monte Carlo should be within reasonable bounds (e.g., ±0.1)
	expectedProb := 0.9 * 0.8
	if math.Abs(sampleRes.Estimate-expectedProb) > 0.1 {
		t.Errorf("Monte Carlo estimate %f too far from expected %f", sampleRes.Estimate, expectedProb)
	}

	// Confidence interval should contain the true value (most of the time)
	if sampleRes.CI95Low > expectedProb || sampleRes.CI95High < expectedProb {
		t.Logf("Warning: true value %f outside CI [%f, %f] (can happen occasionally)",
			expectedProb, sampleRes.CI95Low, sampleRes.CI95High)
	}

	// CI should be reasonable
	if sampleRes.CI95Low < 0 || sampleRes.CI95High > 1 {
		t.Errorf("CI bounds should be in [0,1], got [%f, %f]", sampleRes.CI95Low, sampleRes.CI95High)
	}
}

func TestReachabilityProbabilityQuery_MonteCarlo_DiamondGraph(t *testing.T) {
	g := buildDiamondGraph(t)
	q := ReachabilityProbabilityQuery{Start: "A", End: "D", Mode: MonteCarlo}

	res, err := q.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	sampleRes, ok := res.(result.SampleResult)
	if !ok {
		t.Fatalf("expected SampleResult, got %T", res)
	}

	// Expected: 1 - (1-0.63)*(1-0.48) ≈ 0.8076
	path1 := 0.9 * 0.7
	path2 := 0.8 * 0.6
	expectedProb := 1.0 - (1.0-path1)*(1.0-path2)

	if math.Abs(sampleRes.Estimate-expectedProb) > 0.1 {
		t.Errorf("Monte Carlo estimate %f too far from expected %f", sampleRes.Estimate, expectedProb)
	}
}

func TestReachabilityProbabilityQuery_ContextCancellation(t *testing.T) {
	g := buildComplexGraph(t)
	q := ReachabilityProbabilityQuery{Start: "A", End: "F", Mode: Exact}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := q.Execute(ctx, g)
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
}
