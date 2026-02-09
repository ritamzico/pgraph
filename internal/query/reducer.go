package query

import (
	"fmt"

	"github.com/ritamzico/pgraph/internal/graph"
	"github.com/ritamzico/pgraph/internal/result"
)

type Reducer interface {
	Reduce([]result.Result) (result.Result, error)
}

type MeanProbabilityReducer struct{}

func (r MeanProbabilityReducer) Reduce(results []result.Result) (result.Result, error) {
	var sum float64
	var count int

	for _, res := range results {
		p, ok := res.(result.ProbabilityResult)
		if !ok {
			return nil, fmt.Errorf("expected ProbabilityResult, got %T", res)
		}
		sum += p.Probability
		count++
	}

	return result.ProbabilityResult{
		Probability: sum / float64(count),
	}, nil
}

type BestPathReducer struct{}

func (r BestPathReducer) Reduce(results []result.Result) (result.Result, error) {
	var best graph.Path
	var found bool

	for _, res := range results {
		pr, ok := res.(result.PathResult)
		if !ok {
			return nil, fmt.Errorf("expected PathResult")
		}

		if !found || pr.Path.Probability > best.Probability {
			best = pr.Path
			found = true
		}
	}

	return result.PathResult{Path: best}, nil
}

type MaxProbabilityReducer struct{}

func (r MaxProbabilityReducer) Reduce(results []result.Result) (result.Result, error) {
	maxProb := 0.0

	for _, res := range results {
		pr, ok := res.(result.ProbabilisticResult)
		if !ok {
			return nil, fmt.Errorf("expected ProbabilisticResult, got %T", res)
		}
		if pr.ProbabilityValue() > maxProb {
			maxProb = pr.ProbabilityValue()
		}
	}

	return result.ProbabilityResult{Probability: maxProb}, nil
}

type MinProbabilityReducer struct{}

func (r MinProbabilityReducer) Reduce(results []result.Result) (result.Result, error) {
	minProb := 1.0

	for _, res := range results {
		pr, ok := res.(result.ProbabilisticResult)
		if !ok {
			return nil, fmt.Errorf("expected ProbabilisticResult, got %T", res)
		}
		if pr.ProbabilityValue() < minProb {
			minProb = pr.ProbabilityValue()
		}
	}

	return result.ProbabilityResult{Probability: minProb}, nil
}

type CountAboveThresholdReducer struct {
	Threshold float64
}

func (r CountAboveThresholdReducer) Reduce(results []result.Result) (result.Result, error) {
	count := 0

	for _, res := range results {
		pr, ok := res.(result.ProbabilisticResult)
		if !ok {
			return nil, fmt.Errorf("expected ProbabilisticResult, got %T", res)
		}
		if pr.ProbabilityValue() >= r.Threshold {
			count++
		}
	}

	return result.ProbabilityResult{
		Probability: float64(count) / float64(len(results)),
	}, nil
}
