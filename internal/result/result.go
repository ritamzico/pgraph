package result

type Result interface {
	Kind() Kind
	String() string
}

type Kind int

const (
	PathResultKind Kind = iota
	PathsResultKind
	ProbabilityResultKind
	SampleResultKind
	MultiResultKind
	BooleanResultKind
)

type ProbabilisticResult interface {
	Result
	ProbabilityValue() float64
}
