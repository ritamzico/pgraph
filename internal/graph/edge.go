package graph

type EdgeID string

// Edges are independent Bernoulli random variables
type Edge struct {
	ID          EdgeID
	From, To    NodeID
	Probability float64
	Props       map[string]Value
}
