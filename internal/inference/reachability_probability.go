package inference

import (
	"fmt"
	"math"
	"math/rand/v2"
	"runtime"

	"github.com/ritamzico/pgraph/internal/graph"
	"github.com/ritamzico/pgraph/internal/result"
	"github.com/ritamzico/pgraph/internal/sampling"
)

func ReachabilityProbability(g graph.ProbabilisticGraphModel, start, end graph.NodeID) (float64, error) {
	visited := make(map[graph.NodeID]bool)
	memo := make(map[graph.NodeID]float64)

	return dfsProbabilisticReachability(g, start, end, visited, memo)
}

func ReachabilityProbabilityMonteCarlo(
	g graph.ProbabilisticGraphModel,
	start, end graph.NodeID,
	numSamples int,
	seed uint64,
) (result.SampleResult, error) {
	// TODO: Add importance sampling (if feasible)

	if numSamples <= 0 {
		return result.SampleResult{}, fmt.Errorf("numSamples must be greater than 0")
	}

	numWorkers := min(runtime.GOMAXPROCS(0), numSamples)

	type workerResult struct {
		successes int
		trials    int
		err       error
	}

	results := make(chan workerResult, numWorkers)
	samplesPerWorker := numSamples / numWorkers
	remainder := numSamples % numWorkers

	for w := 0; w < numWorkers; w++ {
		trials := samplesPerWorker
		if w < remainder {
			trials++
		}

		go func(workerID int, trials int) {
			rng := rand.New(rand.NewPCG(
				seed+uint64(workerID),
				(seed^0xda942042e4dd58b5)+uint64(workerID),
			))

			sampler := sampling.IndependentEdgeSampler{Rand: rng}
			successes := 0

			for i := 0; i < trials; i++ {
				sampledWorld, err := sampler.Sample(g)
				if err != nil {
					results <- workerResult{err: err}
					return
				}

				reachable, err := bfsDeterministicReachability(g, start, end, sampledWorld.EdgeMask)
				if err != nil {
					results <- workerResult{err: err}
					return
				}

				if reachable {
					successes++
				}
			}

			results <- workerResult{
				successes: successes,
				trials:    trials,
			}
		}(w, trials)
	}

	totalSuccesses := 0
	totalTrials := 0

	for i := 0; i < numWorkers; i++ {
		r := <-results
		if r.err != nil {
			return result.SampleResult{}, r.err
		}
		totalSuccesses += r.successes
		totalTrials += r.trials
	}

	p := float64(totalSuccesses) / float64(totalTrials)
	variance := p * (1 - p)
	stderr := math.Sqrt(variance / float64(totalTrials))

	return result.SampleResult{
		Estimate:   p,
		NumSamples: numSamples,
		Variance:   variance,
		StdErr:     stderr,
		CI95Low:    p - sampling.CI95ZScore*stderr,
		CI95High:   p + sampling.CI95ZScore*stderr,
	}, nil
}
