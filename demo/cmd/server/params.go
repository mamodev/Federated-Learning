package main

type ParamType string;
var (
	PText ParamType = "text"
	PNumeric ParamType = "numeric"
)

type Param struct	{
	Name string `json:"name"`
	Type ParamType `json:"type"`
	Value any `json:"value"`
}

type ParamFilterOperator string;
var (
	OpEq ParamFilterOperator = "eq"
	OpGt ParamFilterOperator = "gt"
	OpLt ParamFilterOperator = "lt"
	OpGte ParamFilterOperator = "gte"
	OpLte ParamFilterOperator = "lte"
)

type ParamFilter struct {
	Name string
	Type ParamType
	Value any
	Operator ParamFilterOperator
}

func (p Param) MatchFilter(f ParamFilter) bool {

	if p.Type != f.Type	{
		return false
	}

	switch p.Type {
	case PText:
		switch f.Operator {
		case OpEq:
			return p.Value.(string) == f.Value.(string)
		default:
			return false
		}
		
	case PNumeric:
		switch f.Operator {
		case OpEq:
			return p.Value.(int) == f.Value.(int)
		case OpGt:
			return p.Value.(int) > f.Value.(int)
		case OpLt:
			return p.Value.(int) < f.Value.(int)
		case OpGte:
			return p.Value.(int) >= f.Value.(int)
		case OpLte:
			return p.Value.(int) <= f.Value.(int)
		default:
			return false
		}

	default: 
		return false 
	}
}