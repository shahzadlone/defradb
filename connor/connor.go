package connor

import (
	"fmt"

	"github.com/sourcenetwork/defradb/core"
)

// Match is the default method used in Connor to match some data to a
// set of conditions.
func Match(conditions map[FilterKey]interface{}, data core.Doc) (bool, error) {
	return eq(conditions, data)
}

// matchWith can be used to specify the exact operator to use when performing
// a match operation. This is primarily used when building custom operators or
// if you wish to override the behavior of another operator.
func matchWith(op string, conditions, data interface{}) (bool, error) {
	switch op {
	case "_and":
		return and(conditions, data)
	case "_eq":
		return eq(conditions, data)
	case "_ge":
		return ge(conditions, data)
	case "_gt":
		return gt(conditions, data)
	case "_in":
		return in(conditions, data)
	case "_le":
		return le(conditions, data)
	case "_lt":
		return lt(conditions, data)
	case "_ne":
		return ne(conditions, data)
	case "_nin":
		return nin(conditions, data)
	case "_or":
		return or(conditions, data)
	default:
		return false, fmt.Errorf("unknown operator '%s'", op)
	}
}