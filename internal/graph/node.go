package graph

type NodeID string

type Node struct {
	ID    NodeID
	Props map[string]Value
}
