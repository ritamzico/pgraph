package result

import "fmt"

type SampleResult struct {
	Estimate   float64
	NumSamples int
	Variance   float64
	StdErr     float64
	CI95Low    float64
	CI95High   float64
}

func (r SampleResult) Kind() Kind {
	return SampleResultKind
}

func (r SampleResult) ProbabilityValue() float64 {
	return r.Estimate
}

func (r SampleResult) String() string {
	return fmt.Sprintf("Estimate: %.6f (95%% CI: [%.6f, %.6f])\nSamples: %d, Std Error: %.6f",
		r.Estimate, r.CI95Low, r.CI95High, r.NumSamples, r.StdErr)
}
