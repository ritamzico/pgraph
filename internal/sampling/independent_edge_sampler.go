package sampling

import (
	"math/rand/v2"

	"github.com/ritamzico/pgraph/internal/graph"
)

type IndependentEdgeSampler struct {
	Rand *rand.Rand
}

func (s *IndependentEdgeSampler) Sample(g graph.ProbabilisticGraphModel) (*SampledWorld, error) {
	edgeMask := make(map[*graph.Edge]bool)
	for _, edge := range g.GetEdges() {
		edgeMask[edge] = s.Rand.Float64() <= edge.Probability
	}

	return &SampledWorld{EdgeMask: edgeMask}, nil
}
