package graph

type Condition struct {
	ForcedActiveEdges   []*Edge
	ForcedInactiveEdges []*Edge
	ForcedActiveNodes   []NodeID
	ForcedInactiveNodes []NodeID
}
