package dsl

import "fmt"

type SyntaxError struct {
	Kind    string
	Message string
}

func (e SyntaxError) Error() string {
	return fmt.Sprintf("syntax error (%v): %v", e.Kind, e.Message)
}
