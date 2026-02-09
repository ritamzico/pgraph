package query

import "fmt"

type QueryError struct {
	Kind    string
	Message string
}

func (e QueryError) Error() string {
	return fmt.Sprintf("query error (%v): %v", e.Kind, e.Message)
}
