package atom

import (
	"fmt"
	"sort"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func builtinAttr(recv starlark.Value, name string, methods map[string]*starlark.Builtin) (starlark.Value, error) {
	b := methods[name]
	if b == nil {
		return nil, nil // no such method
	}
	return b.BindReceiver(recv), nil
}

func builtinAttrNames(methods map[string]*starlark.Builtin) []string {
	names := make([]string, 0, len(methods))
	for name := range methods {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// threewayCompare interprets a three-way comparison value cmp (-1, 0, +1)
// as a boolean comparison (e.g. x < y).
func threewayCompare(op syntax.Token, cmp int) (bool, error) {
	switch op {
	case syntax.EQL:
		return cmp == 0, nil
	case syntax.NEQ:
		return cmp != 0, nil
	case syntax.LE:
		return cmp <= 0, nil
	case syntax.LT:
		return cmp < 0, nil
	case syntax.GE:
		return cmp >= 0, nil
	case syntax.GT:
		return cmp > 0, nil
	default:
		return false, fmt.Errorf("unexpected comparison operator %s", op)
	}
}
