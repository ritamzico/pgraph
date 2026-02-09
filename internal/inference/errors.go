package inference

import "fmt"

type InferenceError struct {
	Kind string
	Message string
}

func (e InferenceError) Error() string {
	return fmt.Sprintf("inference error (%v): %v", e.Kind, e.Message)
}
