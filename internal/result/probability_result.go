package result

import "fmt"

type ProbabilityResult struct {
	Probability float64
}

func (r ProbabilityResult) Kind() Kind {
	return ProbabilityResultKind
}

func (r ProbabilityResult) ProbabilityValue() float64 {
	return r.Probability
}

func (r ProbabilityResult) String() string {
	return fmt.Sprintf("Probability: %.6f", r.Probability)
}
