package dsl

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var dslLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Keyword", Pattern: `(?i)\b(CREATE|DELETE|NODE|EDGE|FROM|TO|PROB|MAXPATH|TOPK|REACHABILITY|EXACT|MONTECARLO|MULTI|AND|OR|CONDITIONAL|GIVEN|ACTIVE|INACTIVE|THRESHOLD|AGGREGATE|MEAN|MAX|MIN|BESTPATH|COUNTABOVE|K|TRUE|FALSE)\b`},
	{Name: "Float", Pattern: `\d+\.\d+`},
	{Name: "Int", Pattern: `\d+`},
	{Name: "String", Pattern: `"([^"\\]|\\.)*"`},
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
	{Name: "Punct", Pattern: `[(),{}:]`},
	{Name: "Whitespace", Pattern: `\s+`},
})

// Grammar is the top-level AST node.
type Grammar struct {
	Statement *StatementAST `parser:"  @@"`
	Query     *QueryAST     `parser:"| @@"`
}

// StatementAST dispatches on CREATE or DELETE.
type StatementAST struct {
	Create *CreateAST `parser:"\"CREATE\" @@"`
	Delete *DeleteAST `parser:"| \"DELETE\" @@"`
}

// CreateAST dispatches on NODE or EDGE.
type CreateAST struct {
	Node *CreateNodeAST `parser:"\"NODE\" @@"`
	Edge *CreateEdgeAST `parser:"| \"EDGE\" @@"`
}

// CreateNodeAST: comma-separated list of identifiers, with optional properties.
type CreateNodeAST struct {
	IDs   []string   `parser:"@Ident ( \",\" @Ident )*"`
	Props []*PropAST `parser:"( \"{\" @@ ( \",\" @@ )* \"}\" )?"`
}

// CreateEdgeAST: <id> FROM <a> TO <b> PROB <p>, with optional properties.
type CreateEdgeAST struct {
	EdgeID string     `parser:"@Ident"`
	From   string     `parser:"\"FROM\" @Ident"`
	To     string     `parser:"\"TO\" @Ident"`
	Prob   float64    `parser:"\"PROB\" @Float"`
	Props  []*PropAST `parser:"( \"{\" @@ ( \",\" @@ )* \"}\" )?"`
}

// PropAST: <key> : <value>
type PropAST struct {
	Key   string        `parser:"@Ident \":\""`
	Value *PropValueAST `parser:"@@"`
}

// PropValueAST: a typed property value.
type PropValueAST struct {
	Str   *string  `parser:"  @String"`
	Float *float64 `parser:"| @Float"`
	Int   *int64   `parser:"| @Int"`
	True  bool     `parser:"| @\"TRUE\""`
	False bool     `parser:"| @\"FALSE\""`
}

// DeleteAST dispatches on NODE or EDGE.
type DeleteAST struct {
	Node *DeleteNodeAST `parser:"\"NODE\" @@"`
	Edge *DeleteEdgeAST `parser:"| \"EDGE\" @@"`
}

// DeleteNodeAST: comma-separated list of identifiers.
type DeleteNodeAST struct {
	IDs []string `parser:"@Ident ( \",\" @Ident )*"`
}

// DeleteEdgeAST: either "FROM <a> TO <b>" or "<id>".
type DeleteEdgeAST struct {
	FromTo *DeleteEdgeFromToAST `parser:"  @@"`
	ByID   *DeleteEdgeByIDAST   `parser:"| @@"`
}

// DeleteEdgeFromToAST: FROM <a> TO <b>
type DeleteEdgeFromToAST struct {
	From string `parser:"\"FROM\" @Ident"`
	To   string `parser:"\"TO\" @Ident"`
}

// DeleteEdgeByIDAST: <id>
type DeleteEdgeByIDAST struct {
	EdgeID string `parser:"@Ident"`
}

// QueryAST dispatches on the query keyword.
type QueryAST struct {
	Conditional  *ConditionalAST  `parser:"\"CONDITIONAL\" @@"`
	Threshold    *ThresholdAST    `parser:"| \"THRESHOLD\" @@"`
	Aggregate    *AggregateAST    `parser:"| \"AGGREGATE\" @@"`
	MaxPath      *MaxPathAST      `parser:"| \"MAXPATH\" @@"`
	TopK         *TopKAST         `parser:"| \"TOPK\" @@"`
	Reachability *ReachabilityAST `parser:"| \"REACHABILITY\" @@"`
	Multi        *CompositeAST    `parser:"| \"MULTI\" @@"`
	And          *CompositeAST    `parser:"| \"AND\" @@"`
	Or           *CompositeAST    `parser:"| \"OR\" @@"`
}

// MaxPathAST: FROM <a> TO <b>
type MaxPathAST struct {
	From string `parser:"\"FROM\" @Ident"`
	To   string `parser:"\"TO\" @Ident"`
}

// TopKAST: FROM <a> TO <b> K <n>
type TopKAST struct {
	From string `parser:"\"FROM\" @Ident"`
	To   string `parser:"\"TO\" @Ident"`
	K    int    `parser:"\"K\" @Int"`
}

// ReachabilityAST: FROM <a> TO <b> [EXACT|MONTECARLO]
type ReachabilityAST struct {
	From string `parser:"\"FROM\" @Ident"`
	To   string `parser:"\"TO\" @Ident"`
	Mode string `parser:"@( \"EXACT\" | \"MONTECARLO\" )?"`
}

// CompositeAST: ( <query> ( , <query> )* )
type CompositeAST struct {
	Queries []*QueryAST `parser:"\"(\" @@ ( \",\" @@ )* \")\""`
}

// ConditionalAST: GIVEN <conditions> ( <query> )
type ConditionalAST struct {
	Conditions []*ConditionItemAST `parser:"\"GIVEN\" @@ ( \",\" @@ )*"`
	Query      *QueryAST           `parser:"\"(\" @@ \")\""`
}

// ThresholdAST: <threshold> ( <query> )
type ThresholdAST struct {
	Threshold float64   `parser:"@Float"`
	Query     *QueryAST `parser:"\"(\" @@ \")\""`
}

// AggregateAST: <reducer> ( <query> ( , <query> )* )
type AggregateAST struct {
	Reducer *ReducerAST `parser:"@@"`
	Queries []*QueryAST `parser:"\"(\" @@ ( \",\" @@ )* \")\""`
}

// ReducerAST: MEAN | MAX | MIN | BESTPATH | COUNTABOVE <float>
type ReducerAST struct {
	Mean       bool     `parser:"  @\"MEAN\""`
	Max        bool     `parser:"| @\"MAX\""`
	Min        bool     `parser:"| @\"MIN\""`
	BestPath   bool     `parser:"| @\"BESTPATH\""`
	CountAbove *float64 `parser:"| \"COUNTABOVE\" @Float"`
}

// ConditionItemAST: EDGE <id> ACTIVE/INACTIVE  or  NODE <id> ACTIVE/INACTIVE
type ConditionItemAST struct {
	Edge *EdgeConditionAST `parser:"  \"EDGE\" @@"`
	Node *NodeConditionAST `parser:"| \"NODE\" @@"`
}

// EdgeConditionAST: <edgeID> ACTIVE|INACTIVE
type EdgeConditionAST struct {
	EdgeID string `parser:"@Ident"`
	State  string `parser:"@( \"ACTIVE\" | \"INACTIVE\" )"`
}

// NodeConditionAST: <nodeID> ACTIVE|INACTIVE
type NodeConditionAST struct {
	NodeID string `parser:"@Ident"`
	State  string `parser:"@( \"ACTIVE\" | \"INACTIVE\" )"`
}

// Parser singleton built from the grammar.
var dslParser = participle.MustBuild[Grammar](
	participle.Lexer(dslLexer),
	participle.CaseInsensitive("Keyword"),
	participle.Elide("Whitespace"),
)
