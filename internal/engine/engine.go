package engine

import (
	"context"

	"github.com/ritamzico/pgraph/internal/graph"
	"github.com/ritamzico/pgraph/internal/query"
	"github.com/ritamzico/pgraph/internal/result"
)

type InferenceEngine struct {
	Graph graph.ProbabilisticGraphModel
}

func (ie *InferenceEngine) Execute(query query.Query) (result.Result, error) {
	return query.Execute(context.Background(), ie.Graph)
}

func (ie *InferenceEngine) ExecuteWithContext(ctx context.Context, query query.Query) (result.Result, error) {
	return query.Execute(ctx, ie.Graph)
}
