package graph

type ValueKind int

const (
	IntVal ValueKind = iota
	FloatVal
	StringVal
	BoolVal
)

type Value struct {
	Kind ValueKind
	I    int64
	F    float64
	S    string
	B    bool
}
