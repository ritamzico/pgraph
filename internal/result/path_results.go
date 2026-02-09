package result

import (
	"fmt"
	"strings"

	"github.com/ritamzico/pgraph/internal/graph"
)

type PathResult struct {
	Path graph.Path
}

func (r PathResult) Kind() Kind {
	return PathResultKind
}

func (r PathResult) ProbabilityValue() float64 {
	return r.Path.Probability
}

func (r PathResult) String() string {
	return fmt.Sprintf("Path: %s\nProbability: %.6f", formatPath(r.Path.NodeIDs), r.Path.Probability)
}

type PathsResult struct {
	Paths []graph.Path
}

func (r PathsResult) Kind() Kind {
	return PathsResultKind
}

func (r PathsResult) String() string {
	if len(r.Paths) == 0 {
		return "No paths found."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Paths (%d):", len(r.Paths))
	for i, p := range r.Paths {
		fmt.Fprintf(&b, "\n  %d. %s (%.6f)", i+1, formatPath(p.NodeIDs), p.Probability)
	}
	return b.String()
}

func formatPath(nodes []graph.NodeID) string {
	parts := make([]string, len(nodes))
	for i, n := range nodes {
		parts[i] = string(n)
	}
	return strings.Join(parts, " -> ")
}
