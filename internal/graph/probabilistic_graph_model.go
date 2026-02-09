package graph

type ProbabilisticGraphModel interface {
	AddNode(ID NodeID, props map[string]Value) error
	RemoveNode(ID NodeID) error
	GetNodes() []*Node
	ContainsNode(ID NodeID) bool

	AddEdge(edgeID EdgeID, fromID, toID NodeID, prob float64, props map[string]Value) error
	RemoveEdge(fromID, toID NodeID) error
	RemoveEdgeByID(ID EdgeID) error
	GetEdge(fromID, toID NodeID) (*Edge, error)
	GetEdgeByID(id EdgeID) (*Edge, error)
	GetEdges() []*Edge
	ContainsEdge(fromID, toID NodeID) bool
	ContainsEdgeByID(edge EdgeID) bool

	OutgoingEdges(ID NodeID) ([]*Edge, error)
	IncomingEdges(ID NodeID) ([]*Edge, error)

	ApplyCondition(condition Condition) (ProbabilisticGraphModel, error)

	Clone() ProbabilisticGraphModel
}
