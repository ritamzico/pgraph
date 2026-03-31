package inference

import (
	"runtime"
	"slices"

	"github.com/ritamzico/pgraph/internal/graph"
	"github.com/ritamzico/pgraph/internal/result"
)

// SensitivityAnalysis computes the reachability impact of removing each edge
// using exact DFS. Edges are ranked by Delta (Baseline - Without) descending.
func SensitivityAnalysis(
	g graph.ProbabilisticGraphModel,
	start, end graph.NodeID,
) (result.SensitivityResult, error) {
	return sensitivityAnalysis(g, start, end, false, 0)
}

func SensitivityAnalysisMonteCarlo(
	g graph.ProbabilisticGraphModel,
	start, end graph.NodeID,
	seed uint64,
) (result.SensitivityResult, error) {
	return sensitivityAnalysis(g, start, end, true, seed)
}

func sensitivityAnalysis(
	g graph.ProbabilisticGraphModel,
	start, end graph.NodeID,
	useMonteCarlo bool,
	seed uint64,
) (result.SensitivityResult, error) {
	baseline, err := computeReachability(g, start, end, useMonteCarlo, seed)
	if err != nil {
		return result.SensitivityResult{}, err
	}

	edges := g.GetEdges()
	if len(edges) == 0 {
		return result.SensitivityResult{Baseline: baseline}, nil
	}

	type job struct {
		edge      *graph.Edge
		edgeIndex int
	}

	type jobResult struct {
		edgeID  graph.EdgeID
		from    graph.NodeID
		to      graph.NodeID
		prob    float64
		without float64
		err     error
	}

	numEdges := len(edges)
	numWorkers := min(runtime.GOMAXPROCS(0), numEdges)

	jobs := make(chan job, numEdges)
	results := make(chan jobResult, numEdges)

	for range numWorkers {
		go func() {
			for j := range jobs {
				cond := graph.Condition{
					ForcedInactiveEdges: []*graph.Edge{j.edge},
				}
				condGraph, err := g.ApplyCondition(cond)
				if err != nil {
					results <- jobResult{err: err}
					return
				}

				// Offset seed by edgeIndex+1 so per-edge seeds never coincide
				// with the baseline seed (offset 0).
				without, err := computeReachability(
					condGraph, start, end, useMonteCarlo,
					seed+uint64(j.edgeIndex)+1,
				)
				if err != nil {
					results <- jobResult{err: err}
					return
				}

				results <- jobResult{
					edgeID:  j.edge.ID,
					from:    j.edge.From,
					to:      j.edge.To,
					prob:    j.edge.Probability,
					without: without,
				}
			}
		}()
	}

	for i, e := range edges {
		jobs <- job{edge: e, edgeIndex: i}
	}
	close(jobs)

	impacts := make([]result.EdgeImpact, 0, numEdges)
	for range numEdges {
		r := <-results
		if r.err != nil {
			return result.SensitivityResult{}, r.err
		}
		impacts = append(impacts, result.EdgeImpact{
			EdgeID:      r.edgeID,
			From:        r.from,
			To:          r.to,
			Probability: r.prob,
			Without:     r.without,
			Delta:       baseline - r.without,
		})
	}

	slices.SortFunc(impacts, func(a, b result.EdgeImpact) int {
		if a.Delta > b.Delta {
			return -1
		}
		if a.Delta < b.Delta {
			return 1
		}
		return 0
	})

	return result.SensitivityResult{
		Baseline: baseline,
		Impacts:  impacts,
	}, nil
}

// computeReachability dispatches to exact DFS or Monte Carlo reachability.
func computeReachability(
	g graph.ProbabilisticGraphModel,
	start, end graph.NodeID,
	useMonteCarlo bool,
	seed uint64,
) (float64, error) {
	if useMonteCarlo {
		sr, err := ReachabilityProbabilityMonteCarlo(g, start, end, 10000, seed)
		if err != nil {
			return 0, err
		}
		return sr.Estimate, nil
	}
	return ReachabilityProbability(g, start, end)
}
