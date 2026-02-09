package graph

import "testing"

func TestCloneWithEdges(t *testing.T) {
	g := CreateProbAdjListGraph()
	g.AddNode("A", nil)
	g.AddNode("B", nil)
	g.AddEdge("eAB", "A", "B", 0.9, nil)

	cloned := g.Clone()

	if !cloned.ContainsNode("A") {
		t.Error("cloned graph should contain node A")
	}
	if !cloned.ContainsNode("B") {
		t.Error("cloned graph should contain node B")
	}
	if !cloned.ContainsEdgeByID("eAB") {
		t.Error("cloned graph should contain edge eAB")
	}

	// Verify we can remove nodes from cloned graph
	err := cloned.RemoveNode("A")
	if err != nil {
		t.Errorf("RemoveNode failed: %v", err)
	}
}
