package query

import (
	"math"
	"testing"

	"github.com/ritamzico/pgraph/internal/graph"
	"github.com/ritamzico/pgraph/internal/result"
)

// --- MeanProbabilityReducer ---

func TestMeanProbabilityReducer_TwoResults(t *testing.T) {
	r := MeanProbabilityReducer{}
	results := []result.Result{
		result.ProbabilityResult{Probability: 0.8},
		result.ProbabilityResult{Probability: 0.6},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	prob := res.(result.ProbabilityResult).Probability
	if math.Abs(prob-0.7) > 0.0001 {
		t.Errorf("expected 0.7, got %f", prob)
	}
}

func TestMeanProbabilityReducer_SingleResult(t *testing.T) {
	r := MeanProbabilityReducer{}
	results := []result.Result{
		result.ProbabilityResult{Probability: 0.5},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	prob := res.(result.ProbabilityResult).Probability
	if math.Abs(prob-0.5) > 0.0001 {
		t.Errorf("expected 0.5, got %f", prob)
	}
}

func TestMeanProbabilityReducer_TypeMismatch(t *testing.T) {
	r := MeanProbabilityReducer{}
	results := []result.Result{
		result.PathsResult{Paths: nil},
	}

	_, err := r.Reduce(results)
	if err == nil {
		t.Error("expected error for non-ProbabilityResult input")
	}
}

// --- BestPathReducer ---

func TestBestPathReducer_SelectsHighest(t *testing.T) {
	r := BestPathReducer{}
	results := []result.Result{
		result.PathResult{Path: graph.Path{NodeIDs: []graph.NodeID{"A", "C", "D"}, Probability: 0.48}},
		result.PathResult{Path: graph.Path{NodeIDs: []graph.NodeID{"A", "B", "D"}, Probability: 0.63}},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	pathRes := res.(result.PathResult)
	if math.Abs(pathRes.Path.Probability-0.63) > 0.0001 {
		t.Errorf("expected best path prob 0.63, got %f", pathRes.Path.Probability)
	}
}

func TestBestPathReducer_SinglePath(t *testing.T) {
	r := BestPathReducer{}
	results := []result.Result{
		result.PathResult{Path: graph.Path{NodeIDs: []graph.NodeID{"A", "B"}, Probability: 0.9}},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	pathRes := res.(result.PathResult)
	if math.Abs(pathRes.Path.Probability-0.9) > 0.0001 {
		t.Errorf("expected 0.9, got %f", pathRes.Path.Probability)
	}
}

func TestBestPathReducer_TypeMismatch(t *testing.T) {
	r := BestPathReducer{}
	results := []result.Result{
		result.ProbabilityResult{Probability: 0.5},
	}

	_, err := r.Reduce(results)
	if err == nil {
		t.Error("expected error for non-PathResult input")
	}
}

// --- MaxProbabilityReducer ---

func TestMaxProbabilityReducer_SelectsHighest(t *testing.T) {
	r := MaxProbabilityReducer{}
	results := []result.Result{
		result.ProbabilityResult{Probability: 0.3},
		result.ProbabilityResult{Probability: 0.9},
		result.ProbabilityResult{Probability: 0.6},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	prob := res.(result.ProbabilityResult).Probability
	if math.Abs(prob-0.9) > 0.0001 {
		t.Errorf("expected 0.9, got %f", prob)
	}
}

func TestMaxProbabilityReducer_AllEqual(t *testing.T) {
	r := MaxProbabilityReducer{}
	results := []result.Result{
		result.ProbabilityResult{Probability: 0.5},
		result.ProbabilityResult{Probability: 0.5},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	prob := res.(result.ProbabilityResult).Probability
	if math.Abs(prob-0.5) > 0.0001 {
		t.Errorf("expected 0.5, got %f", prob)
	}
}

func TestMaxProbabilityReducer_SingleResult(t *testing.T) {
	r := MaxProbabilityReducer{}
	results := []result.Result{
		result.ProbabilityResult{Probability: 0.7},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	prob := res.(result.ProbabilityResult).Probability
	if math.Abs(prob-0.7) > 0.0001 {
		t.Errorf("expected 0.7, got %f", prob)
	}
}

func TestMaxProbabilityReducer_AcceptsProbabilisticResult(t *testing.T) {
	r := MaxProbabilityReducer{}
	// PathResult also implements ProbabilisticResult
	results := []result.Result{
		result.ProbabilityResult{Probability: 0.5},
		result.PathResult{Path: graph.Path{Probability: 0.8}},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	prob := res.(result.ProbabilityResult).Probability
	if math.Abs(prob-0.8) > 0.0001 {
		t.Errorf("expected 0.8, got %f", prob)
	}
}

func TestMaxProbabilityReducer_TypeMismatch(t *testing.T) {
	r := MaxProbabilityReducer{}
	results := []result.Result{
		result.MultiResult{Results: nil},
	}

	_, err := r.Reduce(results)
	if err == nil {
		t.Error("expected error for non-ProbabilisticResult input")
	}
}

// --- MinProbabilityReducer ---

func TestMinProbabilityReducer_SelectsLowest(t *testing.T) {
	r := MinProbabilityReducer{}
	results := []result.Result{
		result.ProbabilityResult{Probability: 0.3},
		result.ProbabilityResult{Probability: 0.9},
		result.ProbabilityResult{Probability: 0.6},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	prob := res.(result.ProbabilityResult).Probability
	if math.Abs(prob-0.3) > 0.0001 {
		t.Errorf("expected 0.3, got %f", prob)
	}
}

func TestMinProbabilityReducer_AllEqual(t *testing.T) {
	r := MinProbabilityReducer{}
	results := []result.Result{
		result.ProbabilityResult{Probability: 0.7},
		result.ProbabilityResult{Probability: 0.7},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	prob := res.(result.ProbabilityResult).Probability
	if math.Abs(prob-0.7) > 0.0001 {
		t.Errorf("expected 0.7, got %f", prob)
	}
}

func TestMinProbabilityReducer_SingleResult(t *testing.T) {
	r := MinProbabilityReducer{}
	results := []result.Result{
		result.ProbabilityResult{Probability: 0.4},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	prob := res.(result.ProbabilityResult).Probability
	if math.Abs(prob-0.4) > 0.0001 {
		t.Errorf("expected 0.4, got %f", prob)
	}
}

func TestMinProbabilityReducer_AcceptsProbabilisticResult(t *testing.T) {
	r := MinProbabilityReducer{}
	results := []result.Result{
		result.ProbabilityResult{Probability: 0.5},
		result.PathResult{Path: graph.Path{Probability: 0.2}},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	prob := res.(result.ProbabilityResult).Probability
	if math.Abs(prob-0.2) > 0.0001 {
		t.Errorf("expected 0.2, got %f", prob)
	}
}

func TestMinProbabilityReducer_TypeMismatch(t *testing.T) {
	r := MinProbabilityReducer{}
	results := []result.Result{
		result.BooleanResult{Value: true},
	}

	_, err := r.Reduce(results)
	if err == nil {
		t.Error("expected error for non-ProbabilisticResult input")
	}
}

// --- CountAboveThresholdReducer ---

func TestCountAboveThresholdReducer_BasicCounting(t *testing.T) {
	r := CountAboveThresholdReducer{Threshold: 0.5}
	results := []result.Result{
		result.ProbabilityResult{Probability: 0.6},
		result.ProbabilityResult{Probability: 0.4},
		result.ProbabilityResult{Probability: 0.7},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	prob := res.(result.ProbabilityResult).Probability
	// 2 of 3 are above 0.5, so 2/3 = 0.666...
	if math.Abs(prob-2.0/3.0) > 0.0001 {
		t.Errorf("expected 2/3 = %f, got %f", 2.0/3.0, prob)
	}
}

func TestCountAboveThresholdReducer_AllAbove(t *testing.T) {
	r := CountAboveThresholdReducer{Threshold: 0.3}
	results := []result.Result{
		result.ProbabilityResult{Probability: 0.5},
		result.ProbabilityResult{Probability: 0.8},
		result.ProbabilityResult{Probability: 0.9},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	prob := res.(result.ProbabilityResult).Probability
	// All 3 are above 0.3, so 3/3 = 1.0
	if math.Abs(prob-1.0) > 0.0001 {
		t.Errorf("expected 1.0, got %f", prob)
	}
}

func TestCountAboveThresholdReducer_NoneAbove(t *testing.T) {
	r := CountAboveThresholdReducer{Threshold: 0.95}
	results := []result.Result{
		result.ProbabilityResult{Probability: 0.5},
		result.ProbabilityResult{Probability: 0.8},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	prob := res.(result.ProbabilityResult).Probability
	if math.Abs(prob-0.0) > 0.0001 {
		t.Errorf("expected 0.0, got %f", prob)
	}
}

func TestCountAboveThresholdReducer_ExactThreshold(t *testing.T) {
	r := CountAboveThresholdReducer{Threshold: 0.5}
	results := []result.Result{
		result.ProbabilityResult{Probability: 0.5}, // exactly at threshold — should count
		result.ProbabilityResult{Probability: 0.3},
	}

	res, err := r.Reduce(results)
	if err != nil {
		t.Fatalf("Reduce failed: %v", err)
	}

	prob := res.(result.ProbabilityResult).Probability
	// 1 of 2 meets threshold, so 0.5
	if math.Abs(prob-0.5) > 0.0001 {
		t.Errorf("expected 0.5, got %f", prob)
	}
}

func TestCountAboveThresholdReducer_TypeMismatch(t *testing.T) {
	r := CountAboveThresholdReducer{Threshold: 0.5}
	results := []result.Result{
		result.PathsResult{Paths: nil},
	}

	_, err := r.Reduce(results)
	if err == nil {
		t.Error("expected error for non-ProbabilisticResult input")
	}
}
