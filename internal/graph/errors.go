package graph

import "fmt"

type GraphError struct {
	Kind    string
	Message string
}

func (e GraphError) Error() string {
	return fmt.Sprintf("graph error (%v): %v", e.Kind, e.Message)
}

func NodeAlreadyExists(ID NodeID) error {
	return GraphError{
		Kind:    "NodeAlreadyExists",
		Message: fmt.Sprintf("node %v already exists", ID),
	}
}

func NodeDoesNotExist(ID NodeID) error {
	return GraphError{
		Kind:    "NodeDoesNotExists",
		Message: fmt.Sprintf("node %v does not exist", ID),
	}
}

func EdgeAlreadyExists(ID EdgeID) error {
	return GraphError{
		Kind:    "EdgeAlreadyExists",
		Message: fmt.Sprintf("edge %v already exists", ID),
	}
}

func EdgeDoesNotExist(fromID, toID NodeID) error {
	return GraphError{
		Kind:    "EdgeDoesNotExist",
		Message: fmt.Sprintf("edge from %v to %v does not exist", fromID, toID),
	}
}

func EdgeDoesNotExistByID(ID EdgeID) error {
	return GraphError{
		Kind:    "EdgeDoesNotExist",
		Message: fmt.Sprintf("edge %v does not exist", ID),
	}
}
