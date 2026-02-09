package result

import (
	"fmt"
	"strings"
)

type MultiResult struct {
	Results []Result
}

func (r MultiResult) Kind() Kind { return MultiResultKind }

func (r MultiResult) String() string {
	if len(r.Results) == 0 {
		return "No results."
	}
	var b strings.Builder
	for i, sub := range r.Results {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "[%d] %s", i+1, sub.String())
	}
	return b.String()
}

type BooleanResult struct {
	Value bool
}

func (r BooleanResult) Kind() Kind { return BooleanResultKind }

func (r BooleanResult) String() string {
	if r.Value {
		return "Result: true"
	}
	return "Result: false"
}
