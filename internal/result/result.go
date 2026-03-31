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
	SensitivityResultKind
)

type ProbabilisticResult interface {
	Result
	ProbabilityValue() float64
}
