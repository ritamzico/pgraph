package query

import (
	"context"

	"github.com/ritamzico/pgraph/internal/graph"
	"github.com/ritamzico/pgraph/internal/inference"
	"github.com/ritamzico/pgraph/internal/result"
)

type MaxProbabilityPathQuery struct {
	Start, End graph.NodeID
}

func (q MaxProbabilityPathQuery) Execute(ctx context.Context, g graph.ProbabilisticGraphModel) (result.Result, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	path, err := inference.MaxProbabilityPath(g, q.Start, q.End)
	if err != nil {
		return nil, err
	}

	return result.PathResult{
		Path: path,
	}, nil
}

type TopKProbabilityPathsQuery struct {
	Start, End graph.NodeID
	K          int
}

func (q TopKProbabilityPathsQuery) Execute(ctx context.Context, g graph.ProbabilisticGraphModel) (result.Result, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	paths, err := inference.TopKMaxProbabilityPaths(g, q.Start, q.End, q.K)
	if err != nil {
		return nil, err
	}

	return result.PathsResult{
		Paths: paths,
	}, nil
}

type InferenceMode int

const (
	Exact InferenceMode = iota
	MonteCarlo
)

type ReachabilityProbabilityQuery struct {
	Start, End graph.NodeID
	Mode       InferenceMode
	Seed       uint64
}

func (q ReachabilityProbabilityQuery) Execute(ctx context.Context, g graph.ProbabilisticGraphModel) (result.Result, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var probability float64
	var err error

	switch q.Mode {
	case Exact:
		probability, err = inference.ReachabilityProbability(g, q.Start, q.End)
		if err != nil {
			return nil, err
		}

		return result.ProbabilityResult{
			Probability: probability,
		}, nil
	case MonteCarlo:
		sampleResult, err := inference.ReachabilityProbabilityMonteCarlo(g, q.Start, q.End, 10000, q.Seed)
		if err != nil {
			return nil, err
		}

		return sampleResult, nil

	default:
		return nil, QueryError{
			Kind:    "InvalidMode",
			Message: "inference mode should be query.Exact or query.MonteCarlo",
		}
	}
}
