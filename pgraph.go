package pgraph

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ritamzico/pgraph/internal/dsl"
	"github.com/ritamzico/pgraph/internal/graph"
	"github.com/ritamzico/pgraph/internal/result"
	"github.com/ritamzico/pgraph/internal/serialization"
)

type (
	Result            = result.Result
	PathResult        = result.PathResult
	PathsResult       = result.PathsResult
	ProbabilityResult = result.ProbabilityResult
	SampleResult      = result.SampleResult
	MultiResult       = result.MultiResult
	BooleanResult     = result.BooleanResult
)

type PGraph struct {
	Graph  graph.ProbabilisticGraphModel
	parser dsl.Parser
}

func New() *PGraph {
	g := graph.CreateProbAdjListGraph()
	return &PGraph{
		Graph:  g,
		parser: dsl.CreateParser(g),
	}
}

func Load(r io.Reader) (*PGraph, error) {
	g, err := serialization.ReadJSON(r)
	if err != nil {
		return nil, err
	}
	return &PGraph{
		Graph:  g,
		parser: dsl.CreateParser(g),
	}, nil
}

func LoadFile(path string) (*PGraph, error) {
	g, err := serialization.LoadJSON(path)
	if err != nil {
		return nil, err
	}
	return &PGraph{
		Graph:  g,
		parser: dsl.CreateParser(g),
	}, nil
}

func (p *PGraph) Query(dslQuery string) (Result, error) {
	return p.parser.ParseLine(dslQuery)
}

func (p *PGraph) Save(w io.Writer) error {
	return serialization.WriteJSON(p.Graph, w)
}

func (p *PGraph) SaveFile(path string) error {
	return serialization.SaveJSON(p.Graph, path)
}

type jsonResult struct {
	Kind string `json:"kind"`
	Data any    `json:"data"`
}

func MarshalResultJSON(r Result) ([]byte, error) {
	var jr jsonResult
	switch v := r.(type) {
	case result.PathResult:
		jr = jsonResult{Kind: "path", Data: v}
	case result.PathsResult:
		jr = jsonResult{Kind: "paths", Data: v}
	case result.ProbabilityResult:
		jr = jsonResult{Kind: "probability", Data: v}
	case result.SampleResult:
		jr = jsonResult{Kind: "sample", Data: v}
	case result.BooleanResult:
		jr = jsonResult{Kind: "boolean", Data: v}
	case result.MultiResult:
		items := make([]json.RawMessage, len(v.Results))
		for i, sub := range v.Results {
			b, err := MarshalResultJSON(sub)
			if err != nil {
				return nil, err
			}
			items[i] = b
		}
		jr = jsonResult{Kind: "multi", Data: items}
	default:
		jr = jsonResult{Kind: "unknown", Data: fmt.Sprintf("%v", r)}
	}
	return json.Marshal(jr)
}
