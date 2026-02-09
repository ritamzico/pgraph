package sampling

import "github.com/ritamzico/pgraph/internal/graph"

type SampledWorld struct {
	EdgeMask map[*graph.Edge]bool
}

type WorldSampler interface {
	Sample(g graph.ProbabilisticGraphModel) (*SampledWorld, error)
}
