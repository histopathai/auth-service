package query

type FilterOp string

const (
	OpEqual     FilterOp = "=="
	OpNotEqual  FilterOp = "!="
	OpGreater   FilterOp = ">"
	OpLess      FilterOp = "<"
	OpGreaterEq FilterOp = ">="
	OpLessEq    FilterOp = "<="
	OpIn        FilterOp = "in"
	OpNotIn     FilterOp = "not_in"
	OpContains  FilterOp = "contains"
)

type Filter struct {
	Field    string
	Operator FilterOp
	Value    interface{}
}
