package result

import (
	"fmt"
	"strings"

	"github.com/ritamzico/pgraph/internal/graph"
)

type EdgeImpact struct {
	EdgeID      graph.EdgeID
	From        graph.NodeID
	To          graph.NodeID
	Probability float64 // the edge's own Bernoulli probability
	Without     float64 // reachability with this edge forced inactive
	Delta       float64 // Baseline - Without (higher = more critical)
}

type SensitivityResult struct {
	Baseline float64
	Impacts  []EdgeImpact
}

func (r SensitivityResult) Kind() Kind { return SensitivityResultKind }

func (r SensitivityResult) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Baseline reachability: %.6f\n", r.Baseline)

	if len(r.Impacts) == 0 {
		b.WriteString("No edges to analyse.")
		return b.String()
	}

	fmt.Fprintf(&b, "Impact if removed (%d edges, ranked by Δ):", len(r.Impacts))
	for i, imp := range r.Impacts {
		fmt.Fprintf(&b, "\n  %d. %-20s %s -> %s   [p=%.3f]   without=%.6f   Δ=%.6f",
			i+1,
			string(imp.EdgeID),
			string(imp.From),
			string(imp.To),
			imp.Probability,
			imp.Without,
			imp.Delta,
		)
	}
	return b.String()
}
