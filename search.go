package centreon

// Filter represents a search filter that can be built into a JSON-serializable structure.
type Filter interface {
	Build() any
}

// condition represents a single field-operator-value filter.
type condition struct {
	field    string
	operator string
	value    any
}

func (c condition) Build() any {
	return map[string]any{c.field: map[string]any{c.operator: c.value}}
}

// aggregator represents a logical grouping of filters ($and / $or).
type aggregator struct {
	op       string
	children []Filter
}

func (a aggregator) Build() any {
	built := make([]any, len(a.children))
	for i, child := range a.children {
		built[i] = child.Build()
	}
	return map[string]any{a.op: built}
}

// Single-value operators.

// Eq creates an equality filter: field $eq value.
func Eq(field string, value any) Filter {
	return condition{field: field, operator: "$eq", value: value}
}

// Neq creates a not-equal filter: field $neq value.
func Neq(field string, value any) Filter {
	return condition{field: field, operator: "$neq", value: value}
}

// Lt creates a less-than filter: field $lt value.
func Lt(field string, value any) Filter {
	return condition{field: field, operator: "$lt", value: value}
}

// Le creates a less-than-or-equal filter: field $le value.
func Le(field string, value any) Filter {
	return condition{field: field, operator: "$le", value: value}
}

// Gt creates a greater-than filter: field $gt value.
func Gt(field string, value any) Filter {
	return condition{field: field, operator: "$gt", value: value}
}

// Ge creates a greater-than-or-equal filter: field $ge value.
func Ge(field string, value any) Filter {
	return condition{field: field, operator: "$ge", value: value}
}

// Lk creates a like filter: field $lk value.
func Lk(field string, value any) Filter {
	return condition{field: field, operator: "$lk", value: value}
}

// Nk creates a not-like filter: field $nk value.
func Nk(field string, value any) Filter {
	return condition{field: field, operator: "$nk", value: value}
}

// Rg creates a regex filter: field $rg value.
func Rg(field string, value any) Filter {
	return condition{field: field, operator: "$rg", value: value}
}

// Multi-value operators.

// In creates an inclusion filter: field $in [values...].
func In(field string, values ...any) Filter {
	return condition{field: field, operator: "$in", value: values}
}

// Ni creates an exclusion filter: field $ni [values...].
func Ni(field string, values ...any) Filter {
	return condition{field: field, operator: "$ni", value: values}
}

// Logical operators.

// And combines filters with $and.
func And(filters ...Filter) Filter {
	return aggregator{op: "$and", children: filters}
}

// Or combines filters with $or.
func Or(filters ...Filter) Filter {
	return aggregator{op: "$or", children: filters}
}
