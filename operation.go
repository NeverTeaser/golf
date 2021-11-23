package golf

import "reflect"

type Filter string
type SortOrder string

const (
	Equal    Filter = "eq"
	NotEqual Filter = "not_eq"
	Lt       Filter = "lt"
	Lte      Filter = "lte"
	Gt       Filter = "gt"
	Gte      Filter = "gte"
	Like     Filter = "like"
	NotLike  Filter = "not_like"
	In       Filter = "in"
	NotIn    Filter = "not_in"

	// TODO sort
	Desc SortOrder = "desc"
	Asc  SortOrder = "asc"

	querySep = "_"
)

func getSQLOperation(filter Filter) string {
	switch filter {
	case Equal:
		return "="
	case NotEqual:
		return "!="
	case Lt:
		return "<"
	case Lte:
		return "<="
	case Gt:
		return ">"
	case Gte:
		return ">="
	case Like:
		return "LIKE"
	case NotLike:
		return "NOT LIKE"
	case In:
		return "IN"
	case NotIn:
		return "NOT IN"
	default:
		panic("un support filter")
	}
}

type ValueOperation struct {
	Column string
	Filter Filter
	Value  interface{}
}

type OperationWithType struct {
	ValueOperation
	Type reflect.Type
}
