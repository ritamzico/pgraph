package query

import (
	"context"
	"fmt"
	"sync"

	"github.com/ritamzico/pgraph/internal/graph"
	"github.com/ritamzico/pgraph/internal/result"
)

type resultWrapper struct {
	index int
	res   result.Result
	err   error
}

type reducerFunc func([]result.Result) (result.Result, error)

func executeConcurrent(
	ctx context.Context,
	g graph.ProbabilisticGraphModel,
	queries []Query,
	reduce reducerFunc,
) (result.Result, error) {
	if len(queries) == 0 {
		return nil, QueryError{
			Kind:    "InvalidStructure",
			Message: "query requires at least one subquery",
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make([]result.Result, len(queries))
	resCh := make(chan resultWrapper, len(queries))

	var wg sync.WaitGroup
	wg.Add(len(queries))

	for i, q := range queries {
		go func(i int, q Query) {
			defer wg.Done()
			r, err := q.Execute(ctx, g)
			resCh <- resultWrapper{index: i, res: r, err: err}
		}(i, q)
	}

	go func() {
		wg.Wait()
		close(resCh)
	}()

	for rw := range resCh {
		if rw.err != nil {
			cancel()
			return nil, rw.err
		}
		results[rw.index] = rw.res
	}

	return reduce(results)
}

type ConditionalQuery struct {
	Condition graph.Condition
	Inner     Query
}

func (q ConditionalQuery) Execute(ctx context.Context, g graph.ProbabilisticGraphModel) (result.Result, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	conditionedGraph, err := g.ApplyCondition(q.Condition)
	if err != nil {
		return nil, err
	}

	return q.Inner.Execute(ctx, conditionedGraph)
}

type MultiQuery struct {
	Queries []Query
}

func (q MultiQuery) Execute(ctx context.Context, g graph.ProbabilisticGraphModel) (result.Result, error) {
	return executeConcurrent(ctx, g, q.Queries, func(results []result.Result) (result.Result, error) {
		return result.MultiResult{Results: results}, nil
	})
}

type AggregateQuery struct {
	Queries []Query
	Reducer Reducer
}

func (q AggregateQuery) Execute(ctx context.Context, g graph.ProbabilisticGraphModel) (result.Result, error) {
	multiQuery := MultiQuery{Queries: q.Queries}
	queryResults, err := multiQuery.Execute(ctx, g)

	if err != nil {
		return nil, err
	}

	multiResult, ok := queryResults.(result.MultiResult)
	if !ok {
		return nil, QueryError{
			Kind:    "TypeMismatch",
			Message: fmt.Sprintf("inner query expected ProbabilisticResult, got %T", queryResults),
		}
	}

	return q.Reducer.Reduce(multiResult.Results)
}

type SequentialQuery struct {
	First Query
	Then  func(result.Result) (Query, error)
}

func (q SequentialQuery) Execute(ctx context.Context, g graph.ProbabilisticGraphModel) (result.Result, error) {
	firstResult, err := q.First.Execute(ctx, g)
	if err != nil {
		return nil, err
	}

	thenQuery, err := q.Then(firstResult)
	if err != nil {
		return nil, err
	}

	return thenQuery.Execute(ctx, g)
}

type ThresholdQuery struct {
	Inner     Query
	Threshold float64
}

func (q ThresholdQuery) Execute(ctx context.Context, g graph.ProbabilisticGraphModel) (result.Result, error) {
	queryResult, err := q.Inner.Execute(ctx, g)
	if err != nil {
		return nil, err
	}

	probabilisticResult, ok := queryResult.(result.ProbabilisticResult)
	if !ok {
		return nil, QueryError{
			Kind:    "TypeMismatch",
			Message: fmt.Sprintf("inner query expected ProbabilisticResult, got %T", queryResult),
		}
	}

	if (q.Threshold < 0.0) || (q.Threshold > 1.0) {
		return nil, QueryError{
			Kind:    "InvalidParameter",
			Message: fmt.Sprintf("threshold must be between 0 and 1, got %f", q.Threshold),
		}
	}

	return result.BooleanResult{
		Value: probabilisticResult.ProbabilityValue() >= q.Threshold,
	}, nil
}

type AndQuery struct {
	Queries []Query
}

func (q AndQuery) Execute(ctx context.Context, g graph.ProbabilisticGraphModel) (result.Result, error) {
	return executeConcurrent(ctx, g, q.Queries, func(results []result.Result) (result.Result, error) {
		probability := 1.0

		for _, r := range results {
			pr, ok := r.(result.ProbabilisticResult)
			if !ok {
				return nil, QueryError{
					Kind:    "TypeMismatch",
					Message: fmt.Sprintf("inner query expected ProbabilisticResult, got %T", r),
				}
			}
			probability *= pr.ProbabilityValue()
		}

		return result.ProbabilityResult{Probability: probability}, nil
	})
}

type OrQuery struct {
	Queries []Query
}

func (q OrQuery) Execute(ctx context.Context, g graph.ProbabilisticGraphModel) (result.Result, error) {
	return executeConcurrent(ctx, g, q.Queries, func(results []result.Result) (result.Result, error) {
		probability := 1.0

		for _, r := range results {
			pr, ok := r.(result.ProbabilisticResult)
			if !ok {
				return nil, QueryError{
					Kind:    "TypeMismatch",
					Message: fmt.Sprintf("inner query expected ProbabilisticResult, got %T", r),
				}
			}
			probability *= 1.0 - pr.ProbabilityValue()
		}

		return result.ProbabilityResult{Probability: 1.0 - probability}, nil
	})
}
