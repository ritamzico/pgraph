package inference

import (
	"fmt"

	"github.com/ritamzico/pgraph/internal/graph"
)

func equalNodePrefix(a, b []graph.NodeID) bool {
	if len(a) < len(b) {
		return false
	}

	for i := range b {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func pathProbability(g graph.ProbabilisticGraphModel, nodes []graph.NodeID) float64 {
	prob := 1.0
	for i := 0; i < len(nodes)-1; i++ {
		edge, err := g.GetEdge(nodes[i], nodes[i+1])
		if err != nil {
			return 0.0
		}
		prob *= edge.Probability
	}
	return prob
}

// TopKMaxProbabilityPaths finds the top k most probable paths from start to end.
// It uses MaxProbabilityPath and the Yen's K-Shortest Paths algorithm.
func TopKMaxProbabilityPaths(g graph.ProbabilisticGraphModel, start graph.NodeID, end graph.NodeID, k int) ([]graph.Path, error) {
	if k <= 0 {
		return nil, fmt.Errorf("k must be greater than 0")
	}

	var results []graph.Path
	var candidates []graph.Path

	firstPath, err := MaxProbabilityPath(g, start, end)
	if err != nil {
		return nil, err
	}

	if len(firstPath.NodeIDs) == 0 {
		return nil, nil
	}

	results = append(results, firstPath)

	for i := 1; i < k; i++ {
		prevPath := results[i-1]

		for spurIdx := 0; spurIdx < len(prevPath.NodeIDs)-1; spurIdx++ {
			spurNode := prevPath.NodeIDs[spurIdx]
			rootPathNodes := prevPath.NodeIDs[:spurIdx+1]

			gClone := g.Clone()

			// Remove edges that would recreate previous paths
			for _, p := range results {
				if len(p.NodeIDs) > spurIdx &&
					equalNodePrefix(p.NodeIDs, rootPathNodes) {

					from := p.NodeIDs[spurIdx]
					to := p.NodeIDs[spurIdx+1]
					_ = gClone.RemoveEdge(from, to)
				}
			}

			// Spur path
			spurPath, err := MaxProbabilityPath(gClone, spurNode, end)
			if err != nil || len(spurPath.NodeIDs) == 0 {
				continue
			}

			// Combine root + spur (avoid duplicating spurNode)
			fullNodes := append(
				append([]graph.NodeID{}, rootPathNodes[:len(rootPathNodes)-1]...),
				spurPath.NodeIDs...,
			)

			fullProb := pathProbability(g, fullNodes)

			// Check for duplicates in candidates before adding
			isDuplicate := false
			for _, c := range candidates {
				if len(c.NodeIDs) == len(fullNodes) && equalNodePrefix(c.NodeIDs, fullNodes) {
					isDuplicate = true
					break
				}
			}

			if !isDuplicate {
				candidates = append(candidates, graph.Path{
					NodeIDs:     fullNodes,
					Probability: fullProb,
				})
			}
		}

		if len(candidates) == 0 {
			break
		}

		// Pick best candidate
		bestIdx := 0
		for j := 1; j < len(candidates); j++ {
			if candidates[j].Probability > candidates[bestIdx].Probability {
				bestIdx = j
			}
		}

		results = append(results, candidates[bestIdx])

		// Remove chosen candidate
		candidates = append(candidates[:bestIdx], candidates[bestIdx+1:]...)
	}

	return results, nil
}
