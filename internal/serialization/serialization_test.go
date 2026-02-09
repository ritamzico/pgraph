package serialization

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ritamzico/pgraph/internal/graph"
)

func buildGraph(t *testing.T, nodes []nodeDesc, edges []edgeDesc) *graph.ProbabilisticAdjacencyListGraph {
	t.Helper()
	g := graph.CreateProbAdjListGraph()
	for _, n := range nodes {
		if err := g.AddNode(graph.NodeID(n.id), n.props); err != nil {
			t.Fatalf("AddNode(%s): %v", n.id, err)
		}
	}
	for _, e := range edges {
		if err := g.AddEdge(graph.EdgeID(e.id), graph.NodeID(e.from), graph.NodeID(e.to), e.prob, e.props); err != nil {
			t.Fatalf("AddEdge(%s): %v", e.id, err)
		}
	}
	return g
}

type nodeDesc struct {
	id    string
	props map[string]graph.Value
}

type edgeDesc struct {
	id    string
	from  string
	to    string
	prob  float64
	props map[string]graph.Value
}

// roundTrip serializes a graph to JSON and reads it back.
func roundTrip(t *testing.T, g *graph.ProbabilisticAdjacencyListGraph) *graph.ProbabilisticAdjacencyListGraph {
	t.Helper()
	var buf bytes.Buffer
	if err := WriteJSON(g, &buf); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	got, err := ReadJSON(&buf)
	if err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	return got
}

func assertNodeExists(t *testing.T, g *graph.ProbabilisticAdjacencyListGraph, id string) {
	t.Helper()
	if !g.ContainsNode(graph.NodeID(id)) {
		t.Errorf("expected node %q to exist", id)
	}
}

func assertEdgeExists(t *testing.T, g *graph.ProbabilisticAdjacencyListGraph, from, to string, wantProb float64) {
	t.Helper()
	e, err := g.GetEdge(graph.NodeID(from), graph.NodeID(to))
	if err != nil {
		t.Errorf("expected edge %s→%s to exist: %v", from, to, err)
		return
	}
	if math.Abs(e.Probability-wantProb) > 1e-15 {
		t.Errorf("edge %s→%s prob = %v, want %v", from, to, e.Probability, wantProb)
	}
}

func assertNodeProp(t *testing.T, g *graph.ProbabilisticAdjacencyListGraph, nodeID, key string, want graph.Value) {
	t.Helper()
	nodes := g.GetNodes()
	for _, n := range nodes {
		if string(n.ID) == nodeID {
			got, ok := n.Props[key]
			if !ok {
				t.Errorf("node %s: missing prop %q", nodeID, key)
				return
			}
			assertValuesEqual(t, key, got, want)
			return
		}
	}
	t.Errorf("node %s not found", nodeID)
}

func assertEdgeProp(t *testing.T, g *graph.ProbabilisticAdjacencyListGraph, from, to, key string, want graph.Value) {
	t.Helper()
	e, err := g.GetEdge(graph.NodeID(from), graph.NodeID(to))
	if err != nil {
		t.Errorf("edge %s→%s: %v", from, to, err)
		return
	}
	got, ok := e.Props[key]
	if !ok {
		t.Errorf("edge %s→%s: missing prop %q", from, to, key)
		return
	}
	assertValuesEqual(t, key, got, want)
}

func assertValuesEqual(t *testing.T, label string, got, want graph.Value) {
	if got.Kind != want.Kind {
		t.Errorf("prop %s: value kind = %v, want %v", label, got.Kind, want.Kind)
		return
	}

	switch got.Kind {
	case graph.IntVal:
		if got.I != want.I {
			t.Errorf("prop %s: int value = %v, want %v", label, got.I, want.I)
		}
		return
	case graph.FloatVal:
		if math.Abs(got.F-want.F) > 1e-15 {
			t.Errorf("prop %s: float value = %v, want %v", label, got.F, want.F)
		}
		return
	case graph.StringVal:
		if got.S != want.S {
			t.Errorf("prop %s: string value = %q, want %q", label, got.S, want.S)
		}
		return
	case graph.BoolVal:
		if got.B != want.B {
			t.Errorf("prop %s: bool value = %v, want %v", label, got.B, want.B)
		}
		return
	}
}

func TestRoundTripEmptyGraph(t *testing.T) {
	g := graph.CreateProbAdjListGraph()
	got := roundTrip(t, g)
	if len(got.GetNodes()) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(got.GetNodes()))
	}
	if len(got.GetEdges()) != 0 {
		t.Errorf("expected 0 edges, got %d", len(got.GetEdges()))
	}
}

func TestRoundTripNodesOnly(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{
			{id: "a", props: nil},
			{id: "b", props: map[string]graph.Value{}},
			{id: "c", props: map[string]graph.Value{"label": {Kind: graph.StringVal, S: "node-c"}}},
		},
		nil,
	)
	got := roundTrip(t, g)
	if len(got.GetNodes()) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(got.GetNodes()))
	}
	assertNodeExists(t, got, "a")
	assertNodeExists(t, got, "b")
	assertNodeExists(t, got, "c")
	assertNodeProp(t, got, "c", "label", graph.Value{Kind: graph.StringVal, S: "node-c"})
}

func TestRoundTripSimpleGraph(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{{id: "a"}, {id: "b"}, {id: "c"}},
		[]edgeDesc{
			{id: "e1", from: "a", to: "b", prob: 0.9},
			{id: "e2", from: "b", to: "c", prob: 0.5},
		},
	)
	got := roundTrip(t, g)

	if len(got.GetNodes()) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(got.GetNodes()))
	}
	if len(got.GetEdges()) != 2 {
		t.Errorf("expected 2 edges, got %d", len(got.GetEdges()))
	}
	assertEdgeExists(t, got, "a", "b", 0.9)
	assertEdgeExists(t, got, "b", "c", 0.5)
}

func TestRoundTripAllPropertyTypes(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{{
			id: "n1",
			props: map[string]graph.Value{
				"count":   {Kind: graph.IntVal, I: 42},
				"weight":  {Kind: graph.FloatVal, F: 3.14},
				"name":    {Kind: graph.StringVal, S: "hello"},
				"enabled": {Kind: graph.BoolVal, B: true},
			},
		}},
		nil,
	)
	got := roundTrip(t, g)

	assertNodeProp(t, got, "n1", "count", graph.Value{Kind: graph.IntVal, I: 42})
	assertNodeProp(t, got, "n1", "weight", graph.Value{Kind: graph.FloatVal, F: 3.14})
	assertNodeProp(t, got, "n1", "name", graph.Value{Kind: graph.StringVal, S: "hello"})
	assertNodeProp(t, got, "n1", "enabled", graph.Value{Kind: graph.BoolVal, B: true})
}

func TestRoundTripEdgeProperties(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{{id: "a"}, {id: "b"}},
		[]edgeDesc{{
			id: "e1", from: "a", to: "b", prob: 0.75,
			props: map[string]graph.Value{
				"latency": {Kind: graph.IntVal, I: 100},
				"label":   {Kind: graph.StringVal, S: "supply-link"},
			},
		}},
	)
	got := roundTrip(t, g)

	assertEdgeProp(t, got, "a", "b", "latency", graph.Value{Kind: graph.IntVal, I: 100})
	assertEdgeProp(t, got, "a", "b", "label", graph.Value{Kind: graph.StringVal, S: "supply-link"})
}

func TestRoundTripBoundaryProbabilities(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{{id: "a"}, {id: "b"}, {id: "c"}},
		[]edgeDesc{
			{id: "e0", from: "a", to: "b", prob: 0.0},
			{id: "e1", from: "b", to: "c", prob: 1.0},
		},
	)
	got := roundTrip(t, g)

	assertEdgeExists(t, got, "a", "b", 0.0)
	assertEdgeExists(t, got, "b", "c", 1.0)
}

func TestRoundTripPreservesEdgeIDs(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{{id: "x"}, {id: "y"}},
		[]edgeDesc{{id: "my-edge-id", from: "x", to: "y", prob: 0.5}},
	)
	got := roundTrip(t, g)

	if !got.ContainsEdgeByID(graph.EdgeID("my-edge-id")) {
		t.Error("edge ID 'my-edge-id' not preserved after round-trip")
	}
	e, err := got.GetEdgeByID(graph.EdgeID("my-edge-id"))
	if err != nil {
		t.Fatalf("GetEdgeByID: %v", err)
	}
	if e.From != "x" || e.To != "y" {
		t.Errorf("edge endpoints = %s→%s, want x→y", e.From, e.To)
	}
}

func TestRoundTripDisconnectedComponents(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{{id: "a"}, {id: "b"}, {id: "c"}, {id: "d"}},
		[]edgeDesc{
			{id: "e1", from: "a", to: "b", prob: 0.8},
			{id: "e2", from: "c", to: "d", prob: 0.6},
		},
	)
	got := roundTrip(t, g)

	if len(got.GetNodes()) != 4 {
		t.Errorf("expected 4 nodes, got %d", len(got.GetNodes()))
	}
	assertEdgeExists(t, got, "a", "b", 0.8)
	assertEdgeExists(t, got, "c", "d", 0.6)
	if got.ContainsEdge(graph.NodeID("a"), graph.NodeID("c")) {
		t.Error("unexpected edge a→c")
	}
}

func TestRoundTripSelfLoop(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{{id: "a"}},
		[]edgeDesc{{id: "loop", from: "a", to: "a", prob: 0.3}},
	)
	got := roundTrip(t, g)
	assertEdgeExists(t, got, "a", "a", 0.3)
}

func TestRoundTripSpecialCharactersInIDs(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{
			{id: "node with spaces"},
			{id: "node/with/slashes"},
			{id: "node.with.dots"},
			{id: "unicode-日本語"},
		},
		[]edgeDesc{
			{id: "edge with spaces", from: "node with spaces", to: "node/with/slashes", prob: 0.5},
			{id: "edge-日本語", from: "node.with.dots", to: "unicode-日本語", prob: 0.7},
		},
	)
	got := roundTrip(t, g)

	assertNodeExists(t, got, "node with spaces")
	assertNodeExists(t, got, "unicode-日本語")
	assertEdgeExists(t, got, "node with spaces", "node/with/slashes", 0.5)
	assertEdgeExists(t, got, "node.with.dots", "unicode-日本語", 0.7)
}

func TestRoundTripUnicodeInStringProperty(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{{
			id: "n1",
			props: map[string]graph.Value{
				"desc":  {Kind: graph.StringVal, S: "描述"},
				"emoji": {Kind: graph.StringVal, S: "hello 🌍"},
			},
		}},
		nil,
	)
	got := roundTrip(t, g)

	assertNodeProp(t, got, "n1", "desc", graph.Value{Kind: graph.StringVal, S: "描述"})
	assertNodeProp(t, got, "n1", "emoji", graph.Value{Kind: graph.StringVal, S: "hello 🌍"})
}

func TestRoundTripZeroValues(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{{
			id: "n1",
			props: map[string]graph.Value{
				"zero_int":   {Kind: graph.IntVal, I: 0},
				"zero_float": {Kind: graph.FloatVal, F: 0.0},
				"empty_str":  {Kind: graph.StringVal, S: ""},
				"false_bool": {Kind: graph.BoolVal, B: false},
			},
		}},
		nil,
	)
	got := roundTrip(t, g)

	assertNodeProp(t, got, "n1", "zero_int", graph.Value{Kind: graph.IntVal, I: 0})
	assertNodeProp(t, got, "n1", "zero_float", graph.Value{Kind: graph.FloatVal, F: 0.0})
	assertNodeProp(t, got, "n1", "empty_str", graph.Value{Kind: graph.StringVal, S: ""})
	assertNodeProp(t, got, "n1", "false_bool", graph.Value{Kind: graph.BoolVal, B: false})
}

func TestRoundTripNegativeInt(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{{
			id:    "n1",
			props: map[string]graph.Value{"val": {Kind: graph.IntVal, I: -9999}},
		}},
		nil,
	)
	got := roundTrip(t, g)
	assertNodeProp(t, got, "n1", "val", graph.Value{Kind: graph.IntVal, I: -9999})
}

func TestRoundTripLargeInt(t *testing.T) {
	// 2^52 fits exactly in float64 — should survive the round-trip.
	large := int64(1) << 52
	g := buildGraph(t,
		[]nodeDesc{{
			id:    "n1",
			props: map[string]graph.Value{"big": {Kind: graph.IntVal, I: large}},
		}},
		nil,
	)
	got := roundTrip(t, g)
	assertNodeProp(t, got, "n1", "big", graph.Value{Kind: graph.IntVal, I: large})
}

func TestRoundTripSpecialFloats(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{{
			id: "n1",
			props: map[string]graph.Value{
				"tiny":     {Kind: graph.FloatVal, F: math.SmallestNonzeroFloat64},
				"large":    {Kind: graph.FloatVal, F: math.MaxFloat64},
				"negative": {Kind: graph.FloatVal, F: -1.23e100},
			},
		}},
		nil,
	)
	got := roundTrip(t, g)

	assertNodeProp(t, got, "n1", "tiny", graph.Value{Kind: graph.FloatVal, F: math.SmallestNonzeroFloat64})
	assertNodeProp(t, got, "n1", "large", graph.Value{Kind: graph.FloatVal, F: math.MaxFloat64})
	assertNodeProp(t, got, "n1", "negative", graph.Value{Kind: graph.FloatVal, F: -1.23e100})
}

func TestRoundTripManyNodes(t *testing.T) {
	const n = 100
	nodes := make([]nodeDesc, n)
	for i := range n {
		nodes[i] = nodeDesc{id: strings.Repeat("n", i+1)}
	}
	// Chain edges: n1→n2→n3→...
	edges := make([]edgeDesc, n-1)
	for i := range n - 1 {
		edges[i] = edgeDesc{
			id:   strings.Repeat("e", i+1),
			from: nodes[i].id,
			to:   nodes[i+1].id,
			prob: float64(i+1) / float64(n),
		}
	}

	g := buildGraph(t, nodes, edges)
	got := roundTrip(t, g)

	if len(got.GetNodes()) != n {
		t.Errorf("expected %d nodes, got %d", n, len(got.GetNodes()))
	}
	if len(got.GetEdges()) != n-1 {
		t.Errorf("expected %d edges, got %d", n-1, len(got.GetEdges()))
	}
}

func TestRoundTripDiamondGraph(t *testing.T) {
	//   a
	//  / \
	// b   c
	//  \ /
	//   d
	g := buildGraph(t,
		[]nodeDesc{{id: "a"}, {id: "b"}, {id: "c"}, {id: "d"}},
		[]edgeDesc{
			{id: "ab", from: "a", to: "b", prob: 0.9},
			{id: "ac", from: "a", to: "c", prob: 0.8},
			{id: "bd", from: "b", to: "d", prob: 0.7},
			{id: "cd", from: "c", to: "d", prob: 0.6},
		},
	)
	got := roundTrip(t, g)

	if len(got.GetNodes()) != 4 {
		t.Errorf("expected 4 nodes, got %d", len(got.GetNodes()))
	}
	if len(got.GetEdges()) != 4 {
		t.Errorf("expected 4 edges, got %d", len(got.GetEdges()))
	}
	assertEdgeExists(t, got, "a", "b", 0.9)
	assertEdgeExists(t, got, "a", "c", 0.8)
	assertEdgeExists(t, got, "b", "d", 0.7)
	assertEdgeExists(t, got, "c", "d", 0.6)
}

// --- WriteJSON / ReadJSON direct tests ---

func TestWriteJSONProducesValidJSON(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{{id: "a"}, {id: "b"}},
		[]edgeDesc{{id: "e1", from: "a", to: "b", prob: 0.5}},
	)

	var buf bytes.Buffer
	if err := WriteJSON(g, &buf); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	json := buf.String()
	if !strings.Contains(json, `"nodes"`) {
		t.Error("JSON missing 'nodes' key")
	}
	if !strings.Contains(json, `"edges"`) {
		t.Error("JSON missing 'edges' key")
	}
	if !strings.Contains(json, `"probability"`) {
		t.Error("JSON missing 'probability' field")
	}
}

func TestReadJSONEmptyObject(t *testing.T) {
	input := `{}`
	g, err := ReadJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	if len(g.GetNodes()) != 0 || len(g.GetEdges()) != 0 {
		t.Error("expected empty graph from empty JSON object")
	}
}

func TestReadJSONEmptyArrays(t *testing.T) {
	input := `{"nodes": [], "edges": []}`
	g, err := ReadJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	if len(g.GetNodes()) != 0 || len(g.GetEdges()) != 0 {
		t.Error("expected empty graph from empty arrays")
	}
}

func TestReadJSONMinimalNode(t *testing.T) {
	input := `{"nodes": [{"id": "a"}], "edges": []}`
	g, err := ReadJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	assertNodeExists(t, g, "a")
}

func TestReadJSONFullExample(t *testing.T) {
	input := `{
		"nodes": [
			{"id": "a", "props": {"risk": {"kind": "float", "value": 0.8}}},
			{"id": "b"}
		],
		"edges": [
			{"id": "e1", "from": "a", "to": "b", "probability": 0.95, "props": {"type": {"kind": "string", "value": "supply"}}}
		]
	}`
	g, err := ReadJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	assertNodeExists(t, g, "a")
	assertNodeExists(t, g, "b")
	assertEdgeExists(t, g, "a", "b", 0.95)
	assertNodeProp(t, g, "a", "risk", graph.Value{Kind: graph.FloatVal, F: 0.8})
	assertEdgeProp(t, g, "a", "b", "type", graph.Value{Kind: graph.StringVal, S: "supply"})
}

func TestReadJSONInvalidJSON(t *testing.T) {
	inputs := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"bare word", "notjson"},
		{"truncated", `{"nodes": [`},
		{"trailing comma", `{"nodes": [{"id": "a"},]}`},
	}
	for _, tc := range inputs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ReadJSON(strings.NewReader(tc.input))
			if err == nil {
				t.Error("expected error for invalid JSON")
			}
		})
	}
}

func TestReadJSONDuplicateNodeIDs(t *testing.T) {
	input := `{"nodes": [{"id": "a"}, {"id": "a"}], "edges": []}`
	_, err := ReadJSON(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for duplicate node IDs")
	}
}

func TestReadJSONEdgeReferencesNonexistentNode(t *testing.T) {
	input := `{"nodes": [{"id": "a"}], "edges": [{"id": "e1", "from": "a", "to": "b", "probability": 0.5}]}`
	_, err := ReadJSON(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for edge referencing nonexistent node")
	}
}

func TestReadJSONDuplicateEdgeIDs(t *testing.T) {
	input := `{
		"nodes": [{"id": "a"}, {"id": "b"}, {"id": "c"}],
		"edges": [
			{"id": "e1", "from": "a", "to": "b", "probability": 0.5},
			{"id": "e1", "from": "b", "to": "c", "probability": 0.5}
		]
	}`
	_, err := ReadJSON(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for duplicate edge IDs")
	}
}

func TestReadJSONInvalidPropertyType(t *testing.T) {
	input := `{"nodes": [{"id": "a", "props": {"x": {"kind": "int", "value": "not-a-number"}}}], "edges": []}`
	_, err := ReadJSON(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for wrong property value type")
	}
}

func TestReadJSONUnknownValueKind(t *testing.T) {
	input := `{"nodes": [{"id": "a", "props": {"x": {"kind": "complex", "value": 42}}}], "edges": []}`
	_, err := ReadJSON(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for unknown value kind")
	}
}

func TestReadJSONBoolPropertyWrongType(t *testing.T) {
	input := `{"nodes": [{"id": "a", "props": {"x": {"kind": "bool", "value": "yes"}}}], "edges": []}`
	_, err := ReadJSON(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for bool property with string value")
	}
}

func TestReadJSONFloatPropertyWrongType(t *testing.T) {
	input := `{"nodes": [{"id": "a", "props": {"x": {"kind": "float", "value": true}}}], "edges": []}`
	_, err := ReadJSON(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for float property with bool value")
	}
}

func TestReadJSONStringPropertyWrongType(t *testing.T) {
	input := `{"nodes": [{"id": "a", "props": {"x": {"kind": "string", "value": 123}}}], "edges": []}`
	_, err := ReadJSON(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for string property with number value")
	}
}

func TestReadJSONEdgeInvalidPropertyType(t *testing.T) {
	input := `{
		"nodes": [{"id": "a"}, {"id": "b"}],
		"edges": [{"id": "e1", "from": "a", "to": "b", "probability": 0.5, "props": {"x": {"kind": "unknown"}}}]
	}`
	_, err := ReadJSON(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for edge with invalid property type")
	}
}

func TestMarshalValueAllKinds(t *testing.T) {
	cases := []struct {
		name string
		in   graph.Value
		want serializedValue
	}{
		{"int", graph.Value{Kind: graph.IntVal, I: 7}, serializedValue{Kind: "int", Value: int64(7)}},
		{"float", graph.Value{Kind: graph.FloatVal, F: 2.5}, serializedValue{Kind: "float", Value: 2.5}},
		{"string", graph.Value{Kind: graph.StringVal, S: "hi"}, serializedValue{Kind: "string", Value: "hi"}},
		{"bool", graph.Value{Kind: graph.BoolVal, B: true}, serializedValue{Kind: "bool", Value: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := marshalValue(tc.in)
			if got.Kind != tc.want.Kind {
				t.Errorf("Kind = %q, want %q", got.Kind, tc.want.Kind)
			}
		})
	}
}

func TestMarshalValueUnknownKind(t *testing.T) {
	got := marshalValue(graph.Value{Kind: graph.ValueKind(99)})
	if got.Kind != "unknown" {
		t.Errorf("expected 'unknown', got %q", got.Kind)
	}
}

func TestUnmarshalValueRoundTrips(t *testing.T) {
	cases := []struct {
		name string
		sv   serializedValue
		want graph.Value
	}{
		{"int", serializedValue{Kind: "int", Value: float64(42)}, graph.Value{Kind: graph.IntVal, I: 42}},
		{"float", serializedValue{Kind: "float", Value: 3.14}, graph.Value{Kind: graph.FloatVal, F: 3.14}},
		{"string", serializedValue{Kind: "string", Value: "test"}, graph.Value{Kind: graph.StringVal, S: "test"}},
		{"bool_true", serializedValue{Kind: "bool", Value: true}, graph.Value{Kind: graph.BoolVal, B: true}},
		{"bool_false", serializedValue{Kind: "bool", Value: false}, graph.Value{Kind: graph.BoolVal, B: false}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := unmarshalValue(tc.sv)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Kind != tc.want.Kind {
				t.Errorf("Kind = %v, want %v", got.Kind, tc.want.Kind)
			}
		})
	}
}

// --- File I/O tests ---

func TestSaveAndLoadJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "graph.json")

	g := buildGraph(t,
		[]nodeDesc{
			{id: "a", props: map[string]graph.Value{"val": {Kind: graph.IntVal, I: 10}}},
			{id: "b"},
		},
		[]edgeDesc{{id: "e1", from: "a", to: "b", prob: 0.85}},
	)

	if err := SaveJSON(g, path); err != nil {
		t.Fatalf("SaveJSON: %v", err)
	}

	got, err := LoadJSON(path)
	if err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}

	assertNodeExists(t, got, "a")
	assertNodeExists(t, got, "b")
	assertEdgeExists(t, got, "a", "b", 0.85)
	assertNodeProp(t, got, "a", "val", graph.Value{Kind: graph.IntVal, I: 10})
}

func TestLoadJSONNonexistentFile(t *testing.T) {
	_, err := LoadJSON("/nonexistent/path/graph.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestSaveJSONInvalidPath(t *testing.T) {
	g := graph.CreateProbAdjListGraph()
	err := SaveJSON(g, "/nonexistent/dir/graph.json")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestSaveJSONCreatesValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	g := buildGraph(t,
		[]nodeDesc{{id: "a"}},
		nil,
	)
	if err := SaveJSON(g, path); err != nil {
		t.Fatalf("SaveJSON: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, `"id": "a"`) {
		t.Error("file does not contain expected node ID")
	}
}

func TestSaveJSONOverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "graph.json")

	// Write a graph with node "a"
	g1 := buildGraph(t, []nodeDesc{{id: "a"}}, nil)
	if err := SaveJSON(g1, path); err != nil {
		t.Fatalf("SaveJSON (first): %v", err)
	}

	// Overwrite with a graph with node "b"
	g2 := buildGraph(t, []nodeDesc{{id: "b"}}, nil)
	if err := SaveJSON(g2, path); err != nil {
		t.Fatalf("SaveJSON (second): %v", err)
	}

	got, err := LoadJSON(path)
	if err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}

	if got.ContainsNode(graph.NodeID("a")) {
		t.Error("old node 'a' should not be in overwritten graph")
	}
	assertNodeExists(t, got, "b")
}

func TestWriteJSONIsIndented(t *testing.T) {
	g := buildGraph(t, []nodeDesc{{id: "a"}}, nil)
	var buf bytes.Buffer
	if err := WriteJSON(g, &buf); err != nil {
		t.Fatal(err)
	}
	// The encoder uses 2-space indent; check for newlines indicating pretty-printing
	lines := strings.Split(buf.String(), "\n")
	if len(lines) < 3 {
		t.Error("expected indented (multi-line) JSON output")
	}
}

func TestReadJSONIgnoresExtraFields(t *testing.T) {
	input := `{
		"nodes": [{"id": "a", "extra_field": "ignored"}],
		"edges": [],
		"metadata": "also ignored"
	}`
	g, err := ReadJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	assertNodeExists(t, g, "a")
}

func TestRoundTripMultiplePropertiesOnSameNode(t *testing.T) {
	props := map[string]graph.Value{
		"a": {Kind: graph.IntVal, I: 1},
		"b": {Kind: graph.FloatVal, F: 2.0},
		"c": {Kind: graph.StringVal, S: "three"},
		"d": {Kind: graph.BoolVal, B: false},
		"e": {Kind: graph.IntVal, I: -5},
		"f": {Kind: graph.StringVal, S: ""},
	}
	g := buildGraph(t,
		[]nodeDesc{{id: "n", props: props}},
		nil,
	)
	got := roundTrip(t, g)

	for k, want := range props {
		assertNodeProp(t, got, "n", k, want)
	}
}

func TestRoundTripIsolatedNode(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{{id: "lonely"}},
		nil,
	)
	got := roundTrip(t, g)

	assertNodeExists(t, got, "lonely")
	outEdges, _ := got.OutgoingEdges(graph.NodeID("lonely"))
	inEdges, _ := got.IncomingEdges(graph.NodeID("lonely"))
	if len(outEdges) != 0 || len(inEdges) != 0 {
		t.Error("isolated node should have no edges")
	}
}

func TestRoundTripSmallProbability(t *testing.T) {
	g := buildGraph(t,
		[]nodeDesc{{id: "a"}, {id: "b"}},
		[]edgeDesc{{id: "e1", from: "a", to: "b", prob: 1e-15}},
	)
	got := roundTrip(t, g)
	assertEdgeExists(t, got, "a", "b", 1e-15)
}
