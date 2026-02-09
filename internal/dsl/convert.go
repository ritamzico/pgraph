package dsl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ritamzico/pgraph/internal/graph"
	"github.com/ritamzico/pgraph/internal/query"
)

var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func validateIdentifier(name, kind string) error {
	if !validIdentifier.MatchString(name) {
		return SyntaxError{
			Kind:    "InvalidIdentifier",
			Message: fmt.Sprintf("%s identifier %q is invalid: must start with a letter or underscore and contain only letters, digits, and underscores", kind, name),
		}
	}
	return nil
}

func convertGrammar(ast *Grammar, g graph.ProbabilisticGraphModel) (any, error) {
	if ast.Statement != nil {
		return convertStatement(ast.Statement)
	}
	if ast.Query != nil {
		return convertQuery(ast.Query, g)
	}
	return nil, SyntaxError{Kind: "InvalidSyntax", Message: "empty input"}
}

func convertStatement(ast *StatementAST) (Statement, error) {
	if ast.Create != nil {
		return convertCreate(ast.Create)
	}
	return convertDelete(ast.Delete)
}

func convertCreate(ast *CreateAST) (Statement, error) {
	if ast.Node != nil {
		ids := make([]graph.NodeID, len(ast.Node.IDs))
		for i, id := range ast.Node.IDs {
			if err := validateIdentifier(id, "node"); err != nil {
				return nil, err
			}
			ids[i] = graph.NodeID(id)
		}
		return &CreateNodeStatement{
			NodeIDs: ids,
			Props:   convertProps(ast.Node.Props),
		}, nil
	}

	e := ast.Edge
	if err := validateIdentifier(e.EdgeID, "edge"); err != nil {
		return nil, err
	}
	return &CreateEdgeStatement{
		EdgeID: graph.EdgeID(e.EdgeID),
		From:   graph.NodeID(e.From),
		To:     graph.NodeID(e.To),
		Prob:   e.Prob,
		Props:  convertProps(e.Props),
	}, nil
}

func convertProps(props []*PropAST) map[string]graph.Value {
	if len(props) == 0 {
		return nil
	}

	propMap := make(map[string]graph.Value, len(props))

	for _, p := range props {
		var value graph.Value

		switch {
		case p.Value.Str != nil:
			value = graph.Value{Kind: graph.StringVal, S: strings.Trim(*p.Value.Str, "\"")}
		case p.Value.Float != nil:
			value = graph.Value{Kind: graph.FloatVal, F: *p.Value.Float}
		case p.Value.Int != nil:
			value = graph.Value{Kind: graph.IntVal, I: *p.Value.Int}
		case p.Value.True:
			value = graph.Value{Kind: graph.BoolVal, B: true}
		case p.Value.False:
			value = graph.Value{Kind: graph.BoolVal, B: false}
		default:
		}

		propMap[p.Key] = value
	}

	return propMap
}

func convertDelete(ast *DeleteAST) (Statement, error) {
	if ast.Node != nil {
		ids := make([]graph.NodeID, len(ast.Node.IDs))
		for i, id := range ast.Node.IDs {
			ids[i] = graph.NodeID(id)
		}
		return &DeleteNodeStatement{NodeIDs: ids}, nil
	}

	e := ast.Edge
	if e.FromTo != nil {
		return &DeleteEdgeStatement{
			From: graph.NodeID(e.FromTo.From),
			To:   graph.NodeID(e.FromTo.To),
		}, nil
	}
	return &DeleteEdgeByIDStatement{
		EdgeID: graph.EdgeID(e.ByID.EdgeID),
	}, nil
}

func convertQuery(ast *QueryAST, g graph.ProbabilisticGraphModel) (query.Query, error) {
	switch {
	case ast.Conditional != nil:
		return convertConditional(ast.Conditional, g)

	case ast.Threshold != nil:
		return convertThreshold(ast.Threshold, g)

	case ast.Aggregate != nil:
		return convertAggregate(ast.Aggregate, g)

	case ast.MaxPath != nil:
		return query.MaxProbabilityPathQuery{
			Start: graph.NodeID(ast.MaxPath.From),
			End:   graph.NodeID(ast.MaxPath.To),
		}, nil

	case ast.TopK != nil:
		return query.TopKProbabilityPathsQuery{
			Start: graph.NodeID(ast.TopK.From),
			End:   graph.NodeID(ast.TopK.To),
			K:     ast.TopK.K,
		}, nil

	case ast.Reachability != nil:
		r := ast.Reachability
		mode := query.Exact
		if strings.EqualFold(r.Mode, "MONTECARLO") {
			mode = query.MonteCarlo
		}
		return query.ReachabilityProbabilityQuery{
			Start: graph.NodeID(r.From),
			End:   graph.NodeID(r.To),
			Mode:  mode,
		}, nil

	case ast.Multi != nil:
		queries, err := convertComposite(ast.Multi, g)
		if err != nil {
			return nil, err
		}
		return query.MultiQuery{Queries: queries}, nil

	case ast.And != nil:
		queries, err := convertComposite(ast.And, g)
		if err != nil {
			return nil, err
		}
		return query.AndQuery{Queries: queries}, nil

	case ast.Or != nil:
		queries, err := convertComposite(ast.Or, g)
		if err != nil {
			return nil, err
		}
		return query.OrQuery{Queries: queries}, nil

	default:
		return nil, SyntaxError{Kind: "InvalidQuery", Message: fmt.Sprintf("unknown query AST: %+v", ast)}
	}
}

func convertComposite(ast *CompositeAST, g graph.ProbabilisticGraphModel) ([]query.Query, error) {
	queries := make([]query.Query, len(ast.Queries))
	for i, q := range ast.Queries {
		converted, err := convertQuery(q, g)
		if err != nil {
			return nil, err
		}
		queries[i] = converted
	}
	return queries, nil
}

func convertConditional(ast *ConditionalAST, g graph.ProbabilisticGraphModel) (query.Query, error) {
	condition, err := convertCondition(ast.Conditions, g)
	if err != nil {
		return nil, err
	}

	innerQuery, err := convertQuery(ast.Query, g)
	if err != nil {
		return nil, err
	}

	return query.ConditionalQuery{
		Condition: condition,
		Inner:     innerQuery,
	}, nil
}

func convertCondition(items []*ConditionItemAST, g graph.ProbabilisticGraphModel) (graph.Condition, error) {
	var forcedActiveEdges []*graph.Edge
	var forcedInaActiveEdges []*graph.Edge
	var forcedActiveNodes []graph.NodeID
	var forcedInactiveNodes []graph.NodeID

	for _, item := range items {
		switch {
		case item.Edge != nil:
			edgeID := graph.EdgeID(item.Edge.EdgeID)
			edge, err := g.GetEdgeByID(edgeID)

			if err != nil {
				return graph.Condition{}, err
			}

			if item.Edge.State == "ACTIVE" {
				forcedActiveEdges = append(forcedActiveEdges, edge)
			} else {
				forcedInaActiveEdges = append(forcedInaActiveEdges, edge)
			}
		case item.Node != nil:
			nodeID := graph.NodeID(item.Node.NodeID)
			if item.Node.State == "ACTIVE" {
				forcedActiveNodes = append(forcedActiveNodes, nodeID)
			} else {
				forcedInactiveNodes = append(forcedInactiveNodes, nodeID)
			}
		default:
			return graph.Condition{}, SyntaxError{
				Kind:    "InvalidSyntax",
				Message: "empty input",
			}
		}
	}

	return graph.Condition{
		ForcedActiveEdges:   forcedActiveEdges,
		ForcedInactiveEdges: forcedInaActiveEdges,
		ForcedActiveNodes:   forcedInactiveNodes,
		ForcedInactiveNodes: forcedInactiveNodes,
	}, nil
}

func convertThreshold(ast *ThresholdAST, g graph.ProbabilisticGraphModel) (query.Query, error) {
	inner, err := convertQuery(ast.Query, g)
	if err != nil {
		return nil, err
	}

	return query.ThresholdQuery{
		Inner:     inner,
		Threshold: ast.Threshold,
	}, nil
}

func convertAggregate(ast *AggregateAST, g graph.ProbabilisticGraphModel) (query.Query, error) {
	queries := make([]query.Query, len(ast.Queries))
	for i, q := range ast.Queries {
		converted, err := convertQuery(q, g)
		if err != nil {
			return nil, err
		}
		queries[i] = converted
	}

	reducer, err := convertReducer(ast.Reducer)
	if err != nil {
		return nil, err
	}

	return query.AggregateQuery{
		Queries: queries,
		Reducer: reducer,
	}, nil
}

func convertReducer(ast *ReducerAST) (query.Reducer, error) {
	switch {
	case ast.Mean:
		return query.MeanProbabilityReducer{}, nil
	case ast.Max:
		return query.MaxProbabilityReducer{}, nil
	case ast.Min:
		return query.MinProbabilityReducer{}, nil
	case ast.BestPath:
		return query.BestPathReducer{}, nil
	case ast.CountAbove != nil:
		return query.CountAboveThresholdReducer{Threshold: *ast.CountAbove}, nil
	default:
		return nil, SyntaxError{Kind: "InvalidReducer", Message: "unknown reducer"}
	}
}
