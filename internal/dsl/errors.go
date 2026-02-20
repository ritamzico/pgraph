package dsl

import (
	"fmt"
	"strings"
)

// SyntaxError is returned when the DSL input cannot be parsed.
type SyntaxError struct {
	Kind    string
	Message string
}

func (e SyntaxError) Error() string {
	return fmt.Sprintf("syntax error: %v", e.Message)
}

type commandSyntax struct {
	usage   string
	example string
}

var commandHelp = map[string]commandSyntax{
	"create node": {
		usage:   "CREATE NODE <id> [, <id>]* [{ key: value, ... }]",
		example: "CREATE NODE nodeA  OR  CREATE NODE a, b, c",
	},
	"create edge": {
		usage:   "CREATE EDGE <id> FROM <from> TO <to> PROB <probability>",
		example: "CREATE EDGE e1 FROM nodeA TO nodeB PROB 0.9",
	},
	"delete node": {
		usage:   "DELETE NODE <id> [, <id>]*",
		example: "DELETE NODE nodeA",
	},
	"delete edge": {
		usage:   "DELETE EDGE <id>  OR  DELETE EDGE FROM <from> TO <to>",
		example: "DELETE EDGE e1   OR   DELETE EDGE FROM nodeA TO nodeB",
	},
	"maxpath": {
		usage:   "MAXPATH FROM <from> TO <to>",
		example: "MAXPATH FROM nodeA TO nodeB",
	},
	"topk": {
		usage:   "TOPK FROM <from> TO <to> K <n>",
		example: "TOPK FROM nodeA TO nodeB K 3",
	},
	"reachability": {
		usage:   "REACHABILITY FROM <from> TO <to> [EXACT | MONTECARLO]",
		example: "REACHABILITY FROM nodeA TO nodeB EXACT",
	},
	"multi": {
		usage:   "MULTI ( <query>, <query>, ... )",
		example: "MULTI ( MAXPATH FROM a TO b, REACHABILITY FROM c TO d EXACT )",
	},
	"and": {
		usage:   "AND ( <query>, <query>, ... )",
		example: "AND ( REACHABILITY FROM a TO b EXACT, REACHABILITY FROM c TO d EXACT )",
	},
	"or": {
		usage:   "OR ( <query>, <query>, ... )",
		example: "OR ( REACHABILITY FROM a TO b EXACT, REACHABILITY FROM c TO d EXACT )",
	},
	"conditional": {
		usage:   "CONDITIONAL GIVEN [EDGE|NODE] <id> [ACTIVE|INACTIVE] [, ...]* ( <query> )",
		example: "CONDITIONAL GIVEN EDGE e1 INACTIVE ( REACHABILITY FROM a TO b EXACT )",
	},
	"threshold": {
		usage:   "THRESHOLD <probability> ( <query> )",
		example: "THRESHOLD 0.9 ( REACHABILITY FROM a TO b EXACT )",
	},
	"aggregate": {
		usage:   "AGGREGATE [MEAN|MAX|MIN|BESTPATH|COUNTABOVE <float>] ( <query>, ... )",
		example: "AGGREGATE MEAN ( REACHABILITY FROM a TO b EXACT, REACHABILITY FROM c TO d EXACT )",
	},
}

// enrichSyntaxError converts a raw participle parse error into a human-readable SyntaxError.
// It identifies the attempted command from the leading tokens, substitutes internal AST type
// names with readable descriptions, and appends a usage line and example where possible.
func enrichSyntaxError(input string, parseErr error) SyntaxError {
	upper := strings.Fields(strings.ToUpper(input))
	if len(upper) == 0 {
		return SyntaxError{Kind: "InvalidSyntax", Message: "empty input"}
	}

	// Identify the command — try 2-word key first (e.g. "create edge"), then 1-word.
	var cmdLabel string
	var help *commandSyntax
	if len(upper) >= 2 {
		key := strings.ToLower(upper[0]) + " " + strings.ToLower(upper[1])
		if h, ok := commandHelp[key]; ok {
			cmdLabel = strings.ToUpper(key)
			help = &h
		}
	}
	if help == nil {
		key := strings.ToLower(upper[0])
		if h, ok := commandHelp[key]; ok {
			cmdLabel = strings.ToUpper(key)
			help = &h
		}
	}

	// Attempt a specific, targeted diagnosis before falling back to the cleaned raw error.
	specific := specificDiagnostic(upper)
	if specific == "" {
		specific = keywordAsIdentHint(parseErr.Error())
	}

	var b strings.Builder
	if cmdLabel != "" {
		fmt.Fprintf(&b, "%s: ", cmdLabel)
	}
	if specific != "" {
		b.WriteString(specific)
	} else {
		b.WriteString(cleanParticipeError(parseErr.Error()))
	}
	if help != nil {
		fmt.Fprintf(&b, "\n  Usage:   %s", help.usage)
		fmt.Fprintf(&b, "\n  Example: %s", help.example)
	}
	return SyntaxError{Kind: "InvalidSyntax", Message: b.String()}
}

// internalTypeNames maps participle's internal AST struct names and token type names
// to human-readable descriptions.
var internalTypeNames = []struct{ from, to string }{
	{"CreateEdgeAST", `edge ID (e.g. "myEdge")`},
	{"CreateNodeAST", `node ID (e.g. "myNode")`},
	{"DeleteEdgeAST", `edge ID or "FROM <from> TO <to>"`},
	{"DeleteNodeAST", `node ID`},
	{"QueryAST", `query keyword (MAXPATH, TOPK, REACHABILITY, ...)`},
	{"StatementAST", `"CREATE" or "DELETE"`},
	{"CreateAST", `"NODE" or "EDGE"`},
	{"DeleteAST", `"NODE" or "EDGE"`},
	{"MaxPathAST", `FROM <from> TO <to>`},
	{"TopKAST", `FROM <from> TO <to> K <n>`},
	{"ReachabilityAST", `FROM <from> TO <to> [EXACT | MONTECARLO]`},
	{"CompositeAST", `"(" <query> [, <query>]* ")"`},
	{"ConditionalAST", `GIVEN ... ( <query> )`},
	{"ThresholdAST", `<probability> ( <query> )`},
	{"AggregateAST", `<reducer> ( <query>, ... )`},
	{"Grammar", `a valid DSL statement or query`},
	{"<ident>", "identifier"},
}

// keywordAsIdentHint detects the pattern where a DSL keyword appears where an identifier
// was expected, and returns an explanation. It works as a general fallback for any command,
// complementing the per-command diagnostics in specificDiagnostic.
func keywordAsIdentHint(rawErr string) string {
	const prefix = `unexpected token "`
	idx := strings.Index(rawErr, prefix)
	if idx == -1 {
		return ""
	}
	after := rawErr[idx+len(prefix):]
	end := strings.IndexByte(after, '"')
	if end == -1 {
		return ""
	}
	tok := after[:end]
	if dslKeywords[strings.ToUpper(tok)] {
		return fmt.Sprintf("%q is a reserved keyword and cannot be used as an identifier", tok)
	}
	return ""
}

func cleanParticipeError(msg string) string {
	for _, r := range internalTypeNames {
		msg = strings.ReplaceAll(msg, r.from, r.to)
	}
	return msg
}

// dslKeywords is the set of all reserved DSL keywords (uppercased).
var dslKeywords = map[string]bool{
	"CREATE": true, "DELETE": true, "NODE": true, "EDGE": true,
	"FROM": true, "TO": true, "PROB": true,
	"MAXPATH": true, "TOPK": true, "REACHABILITY": true,
	"EXACT": true, "MONTECARLO": true,
	"MULTI": true, "AND": true, "OR": true,
	"CONDITIONAL": true, "GIVEN": true, "ACTIVE": true, "INACTIVE": true,
	"THRESHOLD": true, "AGGREGATE": true,
	"MEAN": true, "MAX": true, "MIN": true, "BESTPATH": true, "COUNTABOVE": true,
	"K": true, "TRUE": true, "FALSE": true,
}

// specificDiagnostic returns a targeted human-readable hint for well-known mistake patterns.
// upper contains the input tokens in uppercase.
func specificDiagnostic(upper []string) string {
	switch upper[0] {
	case "CREATE":
		if len(upper) < 2 {
			return `"NODE" or "EDGE" required after CREATE`
		}
		switch upper[1] {
		case "EDGE":
			return createEdgeDiagnostic(upper[2:])
		case "NODE":
			return createNodeDiagnostic(upper[2:])
		default:
			return fmt.Sprintf("unknown type %q — expected NODE or EDGE", upper[1])
		}
	case "DELETE":
		if len(upper) < 2 {
			return `"NODE" or "EDGE" required after DELETE`
		}
		if upper[1] != "NODE" && upper[1] != "EDGE" {
			return fmt.Sprintf("unknown type %q — expected NODE or EDGE", upper[1])
		}
	case "MAXPATH", "REACHABILITY":
		return fromToDiagnostic(upper[0], upper[1:])
	case "TOPK":
		return topKDiagnostic(upper[1:])
	}
	return ""
}

func createEdgeDiagnostic(rest []string) string {
	if len(rest) == 0 {
		return "edge ID is required after EDGE"
	}
	// The first token after CREATE EDGE must be an identifier, not a keyword.
	if dslKeywords[rest[0]] {
		return fmt.Sprintf("edge ID is required before %q — DSL keywords cannot be used as identifiers", rest[0])
	}
	// Check that the source and destination node IDs are not keywords.
	for i, w := range rest {
		if (w == "FROM" || w == "TO") && i+1 < len(rest) {
			next := rest[i+1]
			if dslKeywords[next] {
				noun := "source"
				if w == "TO" {
					noun = "destination"
				}
				return fmt.Sprintf("%q is a reserved keyword and cannot be used as a %s node ID", strings.ToLower(next), noun)
			}
		}
	}
	// Scan for a PROB clause and validate its value.
	for i, w := range rest {
		if w == "PROB" {
			if i+1 >= len(rest) {
				return "probability value is required after PROB (e.g. PROB 0.9)"
			}
			if !strings.Contains(rest[i+1], ".") {
				return fmt.Sprintf("probability must be a decimal number (e.g. 0.9, not %q)", rest[i+1])
			}
		}
	}
	return ""
}

func createNodeDiagnostic(rest []string) string {
	if len(rest) == 0 {
		return "at least one node ID is required after NODE"
	}
	for _, tok := range rest {
		if dslKeywords[tok] {
			return fmt.Sprintf("%q is a reserved keyword and cannot be used as a node ID", strings.ToLower(tok))
		}
	}
	return ""
}

func fromToDiagnostic(cmd string, rest []string) string {
	if len(rest) == 0 || rest[0] != "FROM" {
		return fmt.Sprintf("FROM keyword is required (e.g. %s FROM nodeA TO nodeB)", cmd)
	}
	// Check that the source node (token after FROM) is not a keyword.
	if len(rest) >= 2 && dslKeywords[rest[1]] && rest[1] != "TO" {
		return fmt.Sprintf("%q is a reserved keyword and cannot be used as a node ID", strings.ToLower(rest[1]))
	}
	for i, w := range rest {
		if w == "TO" {
			if i+1 >= len(rest) {
				return "destination node ID is required after TO"
			}
			// Check that the destination node is not a keyword.
			if dslKeywords[rest[i+1]] {
				return fmt.Sprintf("%q is a reserved keyword and cannot be used as a node ID", strings.ToLower(rest[i+1]))
			}
			return ""
		}
	}
	if len(rest) >= 2 {
		return "TO keyword is required after the source node"
	}
	return "source node ID is required after FROM"
}

func topKDiagnostic(rest []string) string {
	if hint := fromToDiagnostic("TOPK", rest); hint != "" {
		return hint
	}
	for _, w := range rest {
		if w == "K" {
			return ""
		}
	}
	return "K <n> is required at the end (e.g. TOPK FROM nodeA TO nodeB K 3)"
}
