package atom

import "go.starlark.net/starlark"

// for integer
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
		"set": starlark.NewBuiltin("set", intSet),
		"cas": starlark.NewBuiltin("cas", intCAS),
		"add": starlark.NewBuiltin("add", intAdd),
		"sub": starlark.NewBuiltin("sub", intSub),
		"inc": starlark.NewBuiltin("inc", intInc),
		"dec": starlark.NewBuiltin("dec", intDec),
	}
)

func intGet(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicInt)
	return starlark.MakeInt64(recv.val.Load()), nil
}

func intSet(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var value int64
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "value", &value); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicInt)
	recv.val.Store(value)
	return starlark.None, nil
}

func intCAS(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var oldVal, newVal int64
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "old", &oldVal, "new", &newVal); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicInt)
	return starlark.Bool(recv.val.CAS(oldVal, newVal)), nil
}

func intAdd(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var delta int64
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "delta", &delta); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicInt)
	return starlark.MakeInt64(recv.val.Add(delta)), nil
}

func intSub(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var delta int64
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "delta", &delta); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicInt)
	return starlark.MakeInt64(recv.val.Sub(delta)), nil
}

func intInc(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicInt)
	return starlark.MakeInt64(recv.val.Inc()), nil
}

func intDec(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicInt)
	return starlark.MakeInt64(recv.val.Dec()), nil
}
