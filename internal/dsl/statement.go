package dsl

import (
	"github.com/ritamzico/pgraph/internal/graph"
)

type Statement interface {
	Execute(g graph.ProbabilisticGraphModel) error
}

type CreateNodeStatement struct {
	NodeIDs []graph.NodeID
	Props   map[string]graph.Value
}

func (s *CreateNodeStatement) Execute(g graph.ProbabilisticGraphModel) error {
	for _, id := range s.NodeIDs {
		if err := g.AddNode(id, s.Props); err != nil {
			return err
		}
	}
	return nil
}

type DeleteNodeStatement struct {
	NodeIDs []graph.NodeID
}

func (s *DeleteNodeStatement) Execute(g graph.ProbabilisticGraphModel) error {
	for _, id := range s.NodeIDs {
		if err := g.RemoveNode(id); err != nil {
			return err
		}
	}
	return nil
}

type CreateEdgeStatement struct {
	EdgeID graph.EdgeID
	From   graph.NodeID
	To     graph.NodeID
	Prob   float64
	Props  map[string]graph.Value
}

func (s *CreateEdgeStatement) Execute(g graph.ProbabilisticGraphModel) error {
	return g.AddEdge(
		s.EdgeID,
		s.From,
		s.To,
		s.Prob,
		s.Props,
	)
}

type DeleteEdgeStatement struct {
	From graph.NodeID
	To   graph.NodeID
}

func (s *DeleteEdgeStatement) Execute(g graph.ProbabilisticGraphModel) error {
	return g.RemoveEdge(s.From, s.To)
}

type DeleteEdgeByIDStatement struct {
	EdgeID graph.EdgeID
}

func (s *DeleteEdgeByIDStatement) Execute(g graph.ProbabilisticGraphModel) error {
	return g.RemoveEdgeByID(s.EdgeID)
}
