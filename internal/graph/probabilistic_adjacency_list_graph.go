package graph

import (
	"fmt"
	"maps"
	"slices"
)

type ProbabilisticAdjacencyListGraph struct {
	nodeMap map[NodeID]*Node
	edgeMap map[EdgeID]*Edge
	out     map[NodeID]map[NodeID]*Edge
	in      map[NodeID]map[NodeID]*Edge
}

func CreateProbAdjListGraph() *ProbabilisticAdjacencyListGraph {
	graph := &ProbabilisticAdjacencyListGraph{
		nodeMap: make(map[NodeID]*Node),
		edgeMap: make(map[EdgeID]*Edge),
		out:     make(map[NodeID]map[NodeID]*Edge),
		in:      make(map[NodeID]map[NodeID]*Edge),
	}

	return graph
}

func (g *ProbabilisticAdjacencyListGraph) AddNode(ID NodeID, props map[string]Value) error {
	if g.ContainsNode(ID) {
		return NodeAlreadyExists(ID)
	}

	propsCopy := maps.Clone(props)

	newNode := Node{
		ID:    ID,
		Props: propsCopy,
	}

	g.nodeMap[ID] = &newNode
	g.out[ID] = make(map[NodeID]*Edge)
	g.in[ID] = make(map[NodeID]*Edge)

	return nil
}

func (g *ProbabilisticAdjacencyListGraph) RemoveNode(ID NodeID) error {
	if !g.ContainsNode(ID) {
		return NodeDoesNotExist(ID)
	}

	// Get edges before deleting the node, since OutgoingEdges/IncomingEdges check ContainsNode
	outgoingEdges, err := g.OutgoingEdges(ID)
	if err != nil {
		return err
	}

	incomingEdges, err := g.IncomingEdges(ID)
	if err != nil {
		return err
	}

	// Now delete the node
	delete(g.nodeMap, ID)

	// Delete all outgoing edges from edgeMap
	for _, edge := range outgoingEdges {
		delete(g.edgeMap, edge.ID)
	}
	delete(g.out, ID)

	// Delete all incoming edges from edgeMap (may overlap with outgoing)
	for _, edge := range incomingEdges {
		delete(g.edgeMap, edge.ID)
	}
	delete(g.in, ID)

	return nil
}

func (g *ProbabilisticAdjacencyListGraph) GetNodes() []*Node {
	return slices.Collect(maps.Values(g.nodeMap))
}

func (g *ProbabilisticAdjacencyListGraph) ContainsNode(node NodeID) bool {
	_, ok := g.nodeMap[node]
	return ok
}

func (g *ProbabilisticAdjacencyListGraph) AddEdge(edgeID EdgeID, fromID, toID NodeID, prob float64, props map[string]Value) error {
	if g.ContainsEdgeByID(edgeID) {
		return EdgeAlreadyExists(edgeID)
	}

	if !g.ContainsNode(fromID) {
		return NodeDoesNotExist(fromID)
	}

	if !g.ContainsNode(toID) {
		return NodeDoesNotExist(toID)
	}

	if prob < 0 || prob > 1 {
		return GraphError{
			Kind:    "InvalidEdgeProbability",
			Message: "probability must be between 0 and 1",
		}
	}

	propsCopy := maps.Clone(props)

	newEdge := &Edge{
		ID:          edgeID,
		From:        fromID,
		To:          toID,
		Probability: prob,
		Props:       propsCopy,
	}

	g.out[fromID][toID] = newEdge
	g.in[toID][fromID] = newEdge
	g.edgeMap[edgeID] = newEdge

	return nil
}

func (g *ProbabilisticAdjacencyListGraph) RemoveEdge(fromID, toID NodeID) error {
	if !g.ContainsNode(fromID) {
		return NodeDoesNotExist(fromID)
	}

	if !g.ContainsNode(toID) {
		return NodeDoesNotExist(toID)
	}

	if !g.ContainsEdge(fromID, toID) {
		return EdgeDoesNotExist(fromID, toID)
	}

	edgeID := g.out[fromID][toID].ID

	delete(g.out[fromID], toID)
	delete(g.in[toID], fromID)
	delete(g.edgeMap, edgeID)

	return nil
}

func (g *ProbabilisticAdjacencyListGraph) RemoveEdgeByID(edgeID EdgeID) error {
	if !g.ContainsEdgeByID(edgeID) {
		return EdgeDoesNotExistByID(edgeID)
	}

	fromID := g.edgeMap[edgeID].From
	toID := g.edgeMap[edgeID].To

	delete(g.out[fromID], toID)
	delete(g.in[toID], fromID)
	delete(g.edgeMap, edgeID)

	return nil
}

func (g *ProbabilisticAdjacencyListGraph) GetEdge(fromID, toID NodeID) (*Edge, error) {
	if !g.ContainsNode(fromID) {
		return nil, NodeDoesNotExist(fromID)
	}

	if !g.ContainsNode(toID) {
		return nil, NodeDoesNotExist(toID)
	}

	if !g.ContainsEdge(fromID, toID) {
		return nil, EdgeDoesNotExist(fromID, toID)
	}

	return g.out[fromID][toID], nil
}

func (g *ProbabilisticAdjacencyListGraph) GetEdgeByID(id EdgeID) (*Edge, error) {
	edge, ok := g.edgeMap[id]
	if !ok {
		return nil, EdgeDoesNotExistByID(id)
	}
	return edge, nil
}

func (g *ProbabilisticAdjacencyListGraph) GetEdges() []*Edge {
	var allEdges []*Edge
	for _, node := range g.nodeMap {
		edges := slices.Collect(maps.Values(g.out[node.ID]))
		allEdges = append(allEdges, edges...)
	}

	return allEdges
}

func (g *ProbabilisticAdjacencyListGraph) ContainsEdge(fromID, toID NodeID) bool {
	_, ok := g.out[fromID][toID]
	return ok
}

func (g *ProbabilisticAdjacencyListGraph) ContainsEdgeByID(edge EdgeID) bool {
	_, ok := g.edgeMap[edge]
	return ok
}

func (g *ProbabilisticAdjacencyListGraph) OutgoingEdges(ID NodeID) ([]*Edge, error) {
	if !g.ContainsNode(ID) {
		return nil, NodeDoesNotExist(ID)
	}

	return slices.Collect(maps.Values(g.out[ID])), nil
}

func (g *ProbabilisticAdjacencyListGraph) IncomingEdges(ID NodeID) ([]*Edge, error) {
	if !g.ContainsNode(ID) {
		return nil, NodeDoesNotExist(ID)
	}

	return slices.Collect(maps.Values(g.in[ID])), nil
}

func (g *ProbabilisticAdjacencyListGraph) ApplyCondition(condition Condition) (ProbabilisticGraphModel, error) {
	clone := g.Clone().(*ProbabilisticAdjacencyListGraph)

	inactiveNodes := make(map[NodeID]struct{})
	for _, id := range condition.ForcedInactiveNodes {
		inactiveNodes[id] = struct{}{}
	}

	for id := range inactiveNodes {
		if !clone.ContainsNode(id) {
			return nil, GraphError{
				Kind:    "InvalidCondition",
				Message: fmt.Sprintf("node %v from condition does not exist in graph", id),
			}
		}

		for to := range clone.out[id] {
			delete(clone.in[to], id)
		}

		for from := range clone.in[id] {
			delete(clone.out[from], id)
		}

		delete(clone.out, id)
		delete(clone.in, id)
		delete(clone.nodeMap, id)
	}

	for _, edge := range condition.ForcedInactiveEdges {
		from, to := edge.From, edge.To

		if !clone.ContainsNode(from) || !clone.ContainsNode(to) {
			return nil, GraphError{
				Kind:    "InvalidCondition",
				Message: fmt.Sprintf("edge %v from condition does not exist in graph", edge.ID),
			}
		}

		if clone.ContainsEdge(from, to) {
			delete(clone.out[from], to)
			delete(clone.in[to], from)
		}
	}

	return clone, nil
}

func (g *ProbabilisticAdjacencyListGraph) Clone() ProbabilisticGraphModel {
	clone := &ProbabilisticAdjacencyListGraph{
		nodeMap: make(map[NodeID]*Node),
		edgeMap: make(map[EdgeID]*Edge),
		out:     make(map[NodeID]map[NodeID]*Edge),
		in:      make(map[NodeID]map[NodeID]*Edge),
	}

	for id, node := range g.nodeMap {
		newProps := make(map[string]Value)
		maps.Copy(newProps, node.Props)

		clone.nodeMap[id] = &Node{
			ID:    node.ID,
			Props: newProps,
		}

		// Initialize inner maps for all nodes
		clone.out[id] = make(map[NodeID]*Edge)
		clone.in[id] = make(map[NodeID]*Edge)
	}

	for id, edge := range g.edgeMap {
		newProps := make(map[string]Value)
		maps.Copy(newProps, edge.Props)

		clone.edgeMap[id] = &Edge{
			ID:          edge.ID,
			From:        edge.From,
			To:          edge.To,
			Probability: edge.Probability,
			Props:       newProps,
		}
	}

	for from, neighbors := range g.out {
		for to, edge := range neighbors {
			// Use the edge already cloned in edgeMap
			clonedEdge := clone.edgeMap[edge.ID]
			clone.out[from][to] = clonedEdge
			clone.in[to][from] = clonedEdge
		}
	}

	return clone
}
