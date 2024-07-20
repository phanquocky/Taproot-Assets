package common

type InOperator struct {
	Values []any `json:"$in,omitempty"`
}

type OrOperator struct {
	Values []any `json:"$or,omitempty"`
}
