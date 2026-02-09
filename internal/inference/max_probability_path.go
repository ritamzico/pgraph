package inference

import (
	"container/heap"
	"fmt"
	"math"

	"github.com/ritamzico/pgraph/internal/graph"
)

// MaxProbabilityPath finds the path with the highest probability from start to end in a directed graph.
// It uses a modified Dijkstra's algorithm to find the path with the highest probability.
func MaxProbabilityPath(g graph.ProbabilisticGraphModel, start graph.NodeID, end graph.NodeID) (graph.Path, error) {
	if !g.ContainsNode(start) {
		return graph.Path{}, graph.GraphError{
			Kind: "NodeDoesNotExist",
			Message: fmt.Sprintf("start node %v does not exist", start),
		}	
	}

	if !g.ContainsNode(end) {
		return graph.Path{}, graph.GraphError{
			Kind: "NodeDoesNotExist",
			Message: fmt.Sprintf("end node %v does not exist", end),
		}	
	}

	dist := make(map[graph.NodeID]float64)
	prev := make(map[graph.NodeID]graph.NodeID)

	for _, node := range g.GetNodes() {
		dist[node.ID] = math.Inf(1)
	}
	dist[start] = 0.0

	pq := &PriorityQueue{}
	heap.Init(pq)

	heap.Push(pq, &PQItem{
		ID:       start,
		Priority: 0.0,
	})

	for pq.Len() > 0 {
		curr := heap.Pop(pq).(*PQItem)
		u := curr.ID

		if u == end {
			break
		}

		if curr.Priority > dist[u] {
			continue
		}

		outgoingEdges, err := g.OutgoingEdges(u)

		if err != nil {
			return graph.Path{}, err
		}

		for _, edge := range outgoingEdges {
			weight := -math.Log(edge.Probability) // Convert probability to negative log for max-heap
			alt := dist[u] + weight

			if alt < dist[edge.To] {
				dist[edge.To] = alt
				prev[edge.To] = u

				heap.Push(pq, &PQItem{
					ID:       edge.To,
					Priority: alt,
				})
			}
		}
	}

	// No path found
	if math.IsInf(dist[end], 1) {
		return graph.Path{}, nil
	}

	// Reconstruct path
	var pathSlice []graph.NodeID
	for at := end; ; {
		pathSlice = append(pathSlice, at)
		if at == start {
			break
		}
		at = prev[at]
	}

	// Reverse path
	for i, j := 0, len(pathSlice)-1; i < j; i, j = i+1, j-1 {
		pathSlice[i], pathSlice[j] = pathSlice[j], pathSlice[i]
	}

	// Convert back to probability
	prob := math.Exp(-dist[end])

	return graph.Path{NodeIDs: pathSlice, Probability: prob}, nil
}
