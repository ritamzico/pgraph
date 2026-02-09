package query

import (
	"context"

	"github.com/ritamzico/pgraph/internal/graph"
	"github.com/ritamzico/pgraph/internal/result"
)

type Query interface {
	Execute(ctx context.Context, g graph.ProbabilisticGraphModel) (result.Result, error)
}
