package query

import (
	"context"
	"math"
	"testing"

	"github.com/ritamzico/pgraph/internal/graph"
	"github.com/ritamzico/pgraph/internal/result"
)

func TestMultiQuery_TwoReachabilityQueries(t *testing.T) {
	g := buildDiamondGraph(t)

	q1 := ReachabilityProbabilityQuery{Start: "A", End: "B", Mode: Exact}
	q2 := ReachabilityProbabilityQuery{Start: "A", End: "D", Mode: Exact}

	multi := MultiQuery{Queries: []Query{q1, q2}}

	res, err := multi.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	multiRes, ok := res.(result.MultiResult)
	if !ok {
		t.Fatalf("expected MultiResult, got %T", res)
	}

	if len(multiRes.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(multiRes.Results))
	}

	// Verify first result is probability for A->B
	prob1, ok := multiRes.Results[0].(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult for first query, got %T", multiRes.Results[0])
	}
	if math.Abs(prob1.Probability-0.9) > 0.0001 {
		t.Errorf("expected first probability 0.9, got %f", prob1.Probability)
	}

	// Verify second result is probability for A->D
	_, ok = multiRes.Results[1].(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult for second query, got %T", multiRes.Results[1])
	}
}

func TestMultiQuery_MixedQueryTypes(t *testing.T) {
	g := buildDiamondGraph(t)

	q1 := MaxProbabilityPathQuery{Start: "A", End: "D"}
	q2 := ReachabilityProbabilityQuery{Start: "A", End: "D", Mode: Exact}
	q3 := TopKProbabilityPathsQuery{Start: "A", End: "D", K: 2}

	multi := MultiQuery{Queries: []Query{q1, q2, q3}}

	res, err := multi.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	multiRes, ok := res.(result.MultiResult)
	if !ok {
		t.Fatalf("expected MultiResult, got %T", res)
	}

	if len(multiRes.Results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(multiRes.Results))
	}

	// Verify result types
	if _, ok := multiRes.Results[0].(result.PathResult); !ok {
		t.Errorf("expected PathResult for first query, got %T", multiRes.Results[0])
	}
	if _, ok := multiRes.Results[1].(result.ProbabilityResult); !ok {
		t.Errorf("expected ProbabilityResult for second query, got %T", multiRes.Results[1])
	}
	if _, ok := multiRes.Results[2].(result.PathsResult); !ok {
		t.Errorf("expected PathsResult for third query, got %T", multiRes.Results[2])
	}
}

func TestAndQuery_TwoReachabilityQueries(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)

	// Query 1: Reachability A->B (should be 0.9)
	q1 := ReachabilityProbabilityQuery{Start: "A", End: "B", Mode: Exact}
	// Query 2: Reachability B->C (should be 0.8)
	q2 := ReachabilityProbabilityQuery{Start: "B", End: "C", Mode: Exact}

	andQuery := AndQuery{Queries: []Query{q1, q2}}

	res, err := andQuery.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// AND combines probabilities by multiplication: 0.9 * 0.8 = 0.72
	expectedProb := 0.9 * 0.8
	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}

func TestAndQuery_MultipleQueries(t *testing.T) {
	g := buildDiamondGraph(t)

	queries := []Query{
		ReachabilityProbabilityQuery{Start: "A", End: "B", Mode: Exact}, // 0.9
		ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact}, // 0.8
		ReachabilityProbabilityQuery{Start: "B", End: "D", Mode: Exact}, // 0.7
	}

	andQuery := AndQuery{Queries: queries}

	res, err := andQuery.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// AND: 0.9 * 0.8 * 0.7 = 0.504
	expectedProb := 0.9 * 0.8 * 0.7
	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}

func TestOrQuery_TwoReachabilityQueries(t *testing.T) {
	g := buildDiamondGraph(t)

	// Two paths from A to D: via B (0.63) and via C (0.48)
	// We'll query them separately and combine with OR
	q1 := ReachabilityProbabilityQuery{Start: "A", End: "B", Mode: Exact} // 0.9
	q2 := ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact} // 0.8

	orQuery := OrQuery{Queries: []Query{q1, q2}}

	res, err := orQuery.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// OR: 1 - (1-0.9)*(1-0.8) = 1 - 0.1*0.2 = 1 - 0.02 = 0.98
	expectedProb := 1.0 - (1.0-0.9)*(1.0-0.8)
	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}

func TestOrQuery_MultipleQueries(t *testing.T) {
	g := buildLinearGraph(t, 0.6, 0.5)

	queries := []Query{
		ReachabilityProbabilityQuery{Start: "A", End: "B", Mode: Exact}, // 0.6
		ReachabilityProbabilityQuery{Start: "B", End: "C", Mode: Exact}, // 0.5
		ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact}, // 0.3
	}

	orQuery := OrQuery{Queries: queries}

	res, err := orQuery.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// OR: 1 - (1-0.6)*(1-0.5)*(1-0.3) = 1 - 0.4*0.5*0.7 = 1 - 0.14 = 0.86
	expectedProb := 1.0 - (1.0-0.6)*(1.0-0.5)*(1.0-0.3)
	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}

func TestThresholdQuery_AboveThreshold(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)

	innerQuery := ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact}
	thresholdQuery := ThresholdQuery{
		Inner:     innerQuery,
		Threshold: 0.5, // 0.72 > 0.5, should return true
	}

	res, err := thresholdQuery.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	boolRes, ok := res.(result.BooleanResult)
	if !ok {
		t.Fatalf("expected BooleanResult, got %T", res)
	}

	if !boolRes.Value {
		t.Errorf("expected true (0.72 > 0.5), got false")
	}
}

func TestThresholdQuery_BelowThreshold(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)

	innerQuery := ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact}
	thresholdQuery := ThresholdQuery{
		Inner:     innerQuery,
		Threshold: 0.9, // 0.72 < 0.9, should return false
	}

	res, err := thresholdQuery.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	boolRes, ok := res.(result.BooleanResult)
	if !ok {
		t.Fatalf("expected BooleanResult, got %T", res)
	}

	if boolRes.Value {
		t.Errorf("expected false (0.72 < 0.9), got true")
	}
}

func TestThresholdQuery_ExactThreshold(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)

	innerQuery := ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact}
	thresholdQuery := ThresholdQuery{
		Inner:     innerQuery,
		Threshold: 0.72, // 0.72 >= 0.72, should return true
	}

	res, err := thresholdQuery.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	boolRes, ok := res.(result.BooleanResult)
	if !ok {
		t.Fatalf("expected BooleanResult, got %T", res)
	}

	if !boolRes.Value {
		t.Errorf("expected true (0.72 >= 0.72), got false")
	}
}

func TestThresholdQuery_InvalidThreshold_TooLow(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)

	innerQuery := ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact}
	thresholdQuery := ThresholdQuery{
		Inner:     innerQuery,
		Threshold: -0.1, // Invalid
	}

	_, err := thresholdQuery.Execute(context.Background(), g)
	if err == nil {
		t.Error("expected error for threshold < 0")
	}
}

func TestThresholdQuery_InvalidThreshold_TooHigh(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)

	innerQuery := ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact}
	thresholdQuery := ThresholdQuery{
		Inner:     innerQuery,
		Threshold: 1.5, // Invalid
	}

	_, err := thresholdQuery.Execute(context.Background(), g)
	if err == nil {
		t.Error("expected error for threshold > 1")
	}
}

func TestThresholdQuery_NonProbabilisticInnerQuery(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)

	// TopKProbabilityPathsQuery returns PathsResult, which is NOT a ProbabilisticResult
	innerQuery := TopKProbabilityPathsQuery{Start: "A", End: "C", K: 2}
	thresholdQuery := ThresholdQuery{
		Inner:     innerQuery,
		Threshold: 0.5,
	}

	_, err := thresholdQuery.Execute(context.Background(), g)
	if err == nil {
		t.Error("expected error when inner query doesn't return ProbabilisticResult")
	}
}

func TestNestedCompositeQueries(t *testing.T) {
	g := buildDiamondGraph(t)

	// Create nested structure: AND( OR(q1, q2), q3 )
	q1 := ReachabilityProbabilityQuery{Start: "A", End: "B", Mode: Exact} // 0.9
	q2 := ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact} // 0.8
	q3 := ReachabilityProbabilityQuery{Start: "B", End: "D", Mode: Exact} // 0.7

	orQuery := OrQuery{Queries: []Query{q1, q2}}
	andQuery := AndQuery{Queries: []Query{orQuery, q3}}

	res, err := andQuery.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// OR(0.9, 0.8) = 1 - (1-0.9)*(1-0.8) = 0.98
	// AND(0.98, 0.7) = 0.98 * 0.7 = 0.686
	orProb := 1.0 - (1.0-0.9)*(1.0-0.8)
	expectedProb := orProb * 0.7

	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}

func TestThresholdWithCompositeQuery(t *testing.T) {
	g := buildDiamondGraph(t)

	q1 := ReachabilityProbabilityQuery{Start: "A", End: "B", Mode: Exact} // 0.9
	q2 := ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact} // 0.8

	andQuery := AndQuery{Queries: []Query{q1, q2}}
	thresholdQuery := ThresholdQuery{
		Inner:     andQuery,
		Threshold: 0.7, // AND(0.9, 0.8) = 0.72 > 0.7
	}

	res, err := thresholdQuery.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	boolRes, ok := res.(result.BooleanResult)
	if !ok {
		t.Fatalf("expected BooleanResult, got %T", res)
	}

	if !boolRes.Value {
		t.Errorf("expected true (0.72 > 0.7), got false")
	}
}

func TestConditionalQuery_ForcedInactiveEdge(t *testing.T) {
	g := buildDiamondGraph(t)

	edge, err := g.GetEdge("A", "B")
	if err != nil {
		t.Fatalf("Failed to get edge A->B: %v", err)
	}

	inner := ReachabilityProbabilityQuery{Start: "A", End: "D", Mode: Exact}
	condition := graph.Condition{
		ForcedInactiveEdges: []*graph.Edge{edge},
	}

	conditionalQuery := ConditionalQuery{
		Inner:     inner,
		Condition: condition,
	}

	res, err := conditionalQuery.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	// Should only consider A->C->D path (0.8 * 0.6)
	expectedProb := 0.8 * 0.6
	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}
func TestConditionalQuery_ForcedInactiveNode(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)

	node := graph.NodeID("B")

	inner := ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact}
	condition := graph.Condition{
		ForcedInactiveNodes: []graph.NodeID{node},
	}

	conditionalQuery := ConditionalQuery{
		Inner:     inner,
		Condition: condition,
	}

	res, err := conditionalQuery.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	expectedProb := 0.0
	if math.Abs(probRes.Probability-expectedProb) > 0.0001 {
		t.Errorf("expected probability %f, got %f", expectedProb, probRes.Probability)
	}
}

func TestCompositeQuery_EmptyQueryList(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)

	multiQuery := MultiQuery{Queries: []Query{}}

	_, err := multiQuery.Execute(context.Background(), g)
	if err == nil {
		t.Error("expected error for empty query list")
	}
}

func TestCompositeQuery_ContextCancellation(t *testing.T) {
	g := buildComplexGraph(t)

	queries := []Query{
		ReachabilityProbabilityQuery{Start: "A", End: "F", Mode: Exact},
		ReachabilityProbabilityQuery{Start: "B", End: "F", Mode: Exact},
		ReachabilityProbabilityQuery{Start: "C", End: "F", Mode: Exact},
	}

	multiQuery := MultiQuery{Queries: queries}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := multiQuery.Execute(ctx, g)
	// Should handle cancellation gracefully (exact behavior depends on implementation)
	// At minimum, shouldn't hang or panic
	if err == nil {
		t.Log("Note: MultiQuery completed despite cancellation (may be expected if fast)")
	}
}

// --- AggregateQuery tests ---

func TestAggregateQuery_MeanProbabilityReducer(t *testing.T) {
	g := buildDiamondGraph(t)

	queries := []Query{
		ReachabilityProbabilityQuery{Start: "A", End: "B", Mode: Exact}, // 0.9
		ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact}, // 0.8
	}

	agg := AggregateQuery{Queries: queries, Reducer: MeanProbabilityReducer{}}
	res, err := agg.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
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

func TestAggregateQuery_MaxProbabilityReducer(t *testing.T) {
	g := buildDiamondGraph(t)

	queries := []Query{
		ReachabilityProbabilityQuery{Start: "A", End: "B", Mode: Exact}, // 0.9
		ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact}, // 0.8
		ReachabilityProbabilityQuery{Start: "B", End: "D", Mode: Exact}, // 0.7
	}

	agg := AggregateQuery{Queries: queries, Reducer: MaxProbabilityReducer{}}
	res, err := agg.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	if math.Abs(probRes.Probability-0.9) > 0.0001 {
		t.Errorf("expected 0.9, got %f", probRes.Probability)
	}
}

func TestAggregateQuery_MinProbabilityReducer(t *testing.T) {
	g := buildDiamondGraph(t)

	queries := []Query{
		ReachabilityProbabilityQuery{Start: "A", End: "B", Mode: Exact}, // 0.9
		ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact}, // 0.8
		ReachabilityProbabilityQuery{Start: "B", End: "D", Mode: Exact}, // 0.7
	}

	agg := AggregateQuery{Queries: queries, Reducer: MinProbabilityReducer{}}
	res, err := agg.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
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

func TestAggregateQuery_CountAboveThresholdReducer(t *testing.T) {
	g := buildDiamondGraph(t)

	queries := []Query{
		ReachabilityProbabilityQuery{Start: "A", End: "B", Mode: Exact}, // 0.9
		ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact}, // 0.8
		ReachabilityProbabilityQuery{Start: "B", End: "D", Mode: Exact}, // 0.7
	}

	agg := AggregateQuery{
		Queries: queries,
		Reducer: CountAboveThresholdReducer{Threshold: 0.75},
	}
	res, err := agg.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
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

func TestAggregateQuery_BestPathReducer(t *testing.T) {
	g := buildDiamondGraph(t)

	queries := []Query{
		MaxProbabilityPathQuery{Start: "A", End: "D"}, // A→B→D = 0.63
		MaxProbabilityPathQuery{Start: "A", End: "B"}, // A→B = 0.9
	}

	agg := AggregateQuery{Queries: queries, Reducer: BestPathReducer{}}
	res, err := agg.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	pathRes, ok := res.(result.PathResult)
	if !ok {
		t.Fatalf("expected PathResult, got %T", res)
	}

	// Best path is A→B with probability 0.9
	if math.Abs(pathRes.Path.Probability-0.9) > 0.0001 {
		t.Errorf("expected best path prob 0.9, got %f", pathRes.Path.Probability)
	}
}

func TestAggregateQuery_TypeMismatch(t *testing.T) {
	g := buildDiamondGraph(t)

	// MeanProbabilityReducer expects ProbabilityResult, but TopK returns PathsResult
	queries := []Query{
		TopKProbabilityPathsQuery{Start: "A", End: "D", K: 2},
	}

	agg := AggregateQuery{Queries: queries, Reducer: MeanProbabilityReducer{}}
	_, err := agg.Execute(context.Background(), g)
	if err == nil {
		t.Error("expected error for type mismatch between reducer and result")
	}
}

func TestAggregateQuery_SingleQuery(t *testing.T) {
	g := buildLinearGraph(t, 0.9, 0.8)

	queries := []Query{
		ReachabilityProbabilityQuery{Start: "A", End: "C", Mode: Exact}, // 0.72
	}

	agg := AggregateQuery{Queries: queries, Reducer: MaxProbabilityReducer{}}
	res, err := agg.Execute(context.Background(), g)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	probRes, ok := res.(result.ProbabilityResult)
	if !ok {
		t.Fatalf("expected ProbabilityResult, got %T", res)
	}

	if math.Abs(probRes.Probability-0.72) > 0.0001 {
		t.Errorf("expected 0.72, got %f", probRes.Probability)
	}
}
