package inference

import (
	"fmt"

	"github.com/ritamzico/pgraph/internal/graph"
)

func dfsProbabilisticReachability(
	g graph.ProbabilisticGraphModel,
	current, end graph.NodeID,
	visited map[graph.NodeID]bool,
	memo map[graph.NodeID]float64,
) (float64, error) {
	if current == end {
		return 1.0, nil
	}

	if val, ok := memo[current]; ok {
		return val, nil
	}

	if visited[current] {
		return 0.0, nil
	}
	visited[current] = true
	defer delete(visited, current)

	edges, err := g.OutgoingEdges(current)
	if err != nil {
		return 0.0, err
	}

	if len(edges) == 0 {
		memo[current] = 0.0
		return 0.0, nil
	}

	failProb := 1.0

	for _, edge := range edges {
		childProb, err := dfsProbabilisticReachability(g, edge.To, end, visited, memo)

		if err != nil {
			return 0.0, err
		}

		successViaEdge := edge.Probability * childProb
		failProb *= 1.0 - successViaEdge
	}

	result := 1.0 - failProb
	memo[current] = result
	return result, nil
}

func bfsDeterministicReachability(
	g graph.ProbabilisticGraphModel,
	start, end graph.NodeID,
	edgeMask map[*graph.Edge]bool,
) (bool, error) {
	if !g.ContainsNode(start) {
		return false, graph.GraphError{
			Kind: "NodeDoesNotExist",
			Message: fmt.Sprintf("start node %v does not exist", start),
		}
	}

	if !g.ContainsNode(end) {
		return false, graph.GraphError{
			Kind: "NodeDoesNotExist",
			Message: fmt.Sprintf("end node %v does not exist", end),
		}	
	}

	visited := make(map[graph.NodeID]bool)
	queue := []graph.NodeID{start}
	visited[start] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == end {
			return true, nil
		}

		edges, err := g.OutgoingEdges(current)
		if err != nil {
			return false, err
		}

		for _, edge := range edges {
			if !edgeMask[edge] {
				continue
			}

			if !visited[edge.To] {
				visited[edge.To] = true
				queue = append(queue, edge.To)
			}
		}
	}

	return false, nil
}
