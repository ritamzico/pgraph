package dsl

import (
	"fmt"

	"github.com/ritamzico/pgraph/internal/engine"
	"github.com/ritamzico/pgraph/internal/graph"
	"github.com/ritamzico/pgraph/internal/query"
	"github.com/ritamzico/pgraph/internal/result"
)

type Parser struct {
	SessionGraph graph.ProbabilisticGraphModel
	ie           engine.InferenceEngine
}

func CreateParser(baseGraph graph.ProbabilisticGraphModel) Parser {
	clonedGraph := baseGraph.Clone()

	return Parser{
		SessionGraph: clonedGraph,
		ie:           engine.InferenceEngine{Graph: clonedGraph},
	}
}

func (p Parser) ParseLine(input string) (result.Result, error) {
	ast, err := dslParser.ParseString("", input)
	if err != nil {
		return nil, enrichSyntaxError(input, err)
	}

	node, err := convertGrammar(ast, p.SessionGraph)
	if err != nil {
		return nil, err
	}

	switch n := node.(type) {
	case Statement:
		return nil, n.Execute(p.SessionGraph)

	case query.Query:
		return p.ie.Execute(n)

	default:
		return nil, fmt.Errorf("internal error: unknown AST node %T", n)
	}
}
