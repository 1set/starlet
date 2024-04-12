package atom

import "go.starlark.net/starlark"

/*
AtomicInt.get() -> int
AtomicInt.set(value: int)
AtomicInt.cas(old: int, new: int) -> bool
AtomicInt.add(delta: int) -> int
AtomicInt.sub(delta: int) -> int
AtomicInt.inc() -> int
AtomicInt.dec() -> int
*/

var (
	intMethods = map[string]*starlark.Builtin{
		"get": starlark.NewBuiltin("get", intGet),
	}
)

func intGet(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicInt)
	return starlark.MakeInt64(recv.val.Load()), nil
}
