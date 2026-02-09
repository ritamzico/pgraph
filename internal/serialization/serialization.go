package serialization

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ritamzico/pgraph/internal/graph"
)

type serializedValue struct {
	Kind  string `json:"kind"`
	Value any    `json:"value,omitempty"`
}

type serializedNode struct {
	ID    string                     `json:"id"`
	Props map[string]serializedValue `json:"props,omitempty"`
}

type serializedEdge struct {
	ID          string                     `json:"id"`
	From        string                     `json:"from"`
	To          string                     `json:"to"`
	Probability float64                    `json:"probability"`
	Props       map[string]serializedValue `json:"props,omitempty"`
}

type serializedGraph struct {
	Nodes []serializedNode `json:"nodes"`
	Edges []serializedEdge `json:"edges"`
}

func marshalValue(v graph.Value) serializedValue {
	switch v.Kind {
	case graph.IntVal:
		return serializedValue{Kind: "int", Value: v.I}
	case graph.FloatVal:
		return serializedValue{Kind: "float", Value: v.F}
	case graph.StringVal:
		return serializedValue{Kind: "string", Value: v.S}
	case graph.BoolVal:
		return serializedValue{Kind: "bool", Value: v.B}
	default:
		return serializedValue{Kind: "unknown"}
	}
}

func unmarshalValue(sv serializedValue) (graph.Value, error) {
	switch sv.Kind {
	case "int":
		f, ok := sv.Value.(float64)
		if !ok {
			return graph.Value{}, fmt.Errorf("expected number for int, got %T", sv.Value)
		}
		return graph.Value{
			Kind: graph.IntVal,
			I:    int64(f),
		}, nil

	case "float":
		f, ok := sv.Value.(float64)
		if !ok {
			return graph.Value{}, fmt.Errorf("expected number for float, got %T", sv.Value)
		}
		return graph.Value{
			Kind: graph.FloatVal,
			F:    f,
		}, nil

	case "string":
		s, ok := sv.Value.(string)
		if !ok {
			return graph.Value{}, fmt.Errorf("expected string, got %T", sv.Value)
		}
		return graph.Value{
			Kind: graph.StringVal,
			S:    s,
		}, nil

	case "bool":
		b, ok := sv.Value.(bool)
		if !ok {
			return graph.Value{}, fmt.Errorf("expected bool, got %T", sv.Value)
		}
		return graph.Value{
			Kind: graph.BoolVal,
			B:    b,
		}, nil

	default:
		return graph.Value{}, fmt.Errorf("unknown serialized value kind %q", sv.Kind)
	}
}

func toSerializedGraph(g graph.ProbabilisticGraphModel) serializedGraph {
	nodes := g.GetNodes()
	edges := g.GetEdges()

	sNodes := make([]serializedNode, 0, len(nodes))
	for _, n := range nodes {
		sProps := make(map[string]serializedValue, len(n.Props))
		for k, v := range n.Props {
			sProps[k] = marshalValue(v)
		}
		sNodes = append(sNodes, serializedNode{ID: string(n.ID), Props: sProps})
	}

	sEdges := make([]serializedEdge, 0, len(edges))
	for _, e := range edges {
		sProps := make(map[string]serializedValue, len(e.Props))
		for k, v := range e.Props {
			sProps[k] = marshalValue(v)
		}
		sEdges = append(sEdges, serializedEdge{
			ID:          string(e.ID),
			From:        string(e.From),
			To:          string(e.To),
			Probability: e.Probability,
			Props:       sProps,
		})
	}

	return serializedGraph{Nodes: sNodes, Edges: sEdges}
}

func fromSerializedGraph(sg serializedGraph) (*graph.ProbabilisticAdjacencyListGraph, error) {
	g := graph.CreateProbAdjListGraph()

	for _, sn := range sg.Nodes {
		props := make(map[string]graph.Value, len(sn.Props))
		for k, sv := range sn.Props {
			v, err := unmarshalValue(sv)
			if err != nil {
				return nil, fmt.Errorf("node %s prop %s: %w", sn.ID, k, err)
			}
			props[k] = v
		}
		if err := g.AddNode(graph.NodeID(sn.ID), props); err != nil {
			return nil, fmt.Errorf("adding node %s: %w", sn.ID, err)
		}
	}

	for _, se := range sg.Edges {
		props := make(map[string]graph.Value, len(se.Props))
		for k, sv := range se.Props {
			v, err := unmarshalValue(sv)
			if err != nil {
				return nil, fmt.Errorf("edge %s prop %s: %w", se.ID, k, err)
			}
			props[k] = v
		}
		if err := g.AddEdge(
			graph.EdgeID(se.ID),
			graph.NodeID(se.From),
			graph.NodeID(se.To),
			se.Probability,
			props,
		); err != nil {
			return nil, fmt.Errorf("adding edge %s: %w", se.ID, err)
		}
	}

	return g, nil
}

// WriteJSON encodes a graph to JSON and writes it to w.
func WriteJSON(g graph.ProbabilisticGraphModel, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(toSerializedGraph(g))
}

// ReadJSON decodes a graph from JSON read from r.
func ReadJSON(r io.Reader) (*graph.ProbabilisticAdjacencyListGraph, error) {
	var sg serializedGraph
	if err := json.NewDecoder(r).Decode(&sg); err != nil {
		return nil, fmt.Errorf("decoding graph JSON: %w", err)
	}
	return fromSerializedGraph(sg)
}

// SaveJSON writes a graph to a JSON file at path.
func SaveJSON(g graph.ProbabilisticGraphModel, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", path, err)
	}
	defer f.Close()
	return WriteJSON(g, f)
}

// LoadJSON reads a graph from a JSON file at path.
func LoadJSON(path string) (*graph.ProbabilisticAdjacencyListGraph, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", path, err)
	}
	defer f.Close()
	return ReadJSON(f)
}
