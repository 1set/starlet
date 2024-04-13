package atom

import (
	itn "github.com/1set/starlet/internal"
	"go.starlark.net/starlark"
)

// for integer
/*
new_int(value: int) -> AtomicInt

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

// for float
/*
new_float(value: float) -> AtomicFloat

AtomicFloat.get() -> float
AtomicFloat.set(value: float)
AtomicFloat.cas(old: float, new: float) -> bool
AtomicFloat.add(delta: float) -> float
AtomicFloat.sub(delta: float) -> float
*/

var (
	floatMethods = map[string]*starlark.Builtin{
		"get": starlark.NewBuiltin("get", floatGet),
		"set": starlark.NewBuiltin("set", floatSet),
		"cas": starlark.NewBuiltin("cas", floatCAS),
		"add": starlark.NewBuiltin("add", floatAdd),
		"sub": starlark.NewBuiltin("sub", floatSub),
	}
)

func floatGet(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicFloat)
	return starlark.Float(recv.val.Load()), nil
}

func floatSet(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var value itn.FloatOrInt
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "value", &value); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicFloat)
	recv.val.Store(value.GoFloat())
	return starlark.None, nil
}

func floatCAS(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var oldVal, newVal itn.FloatOrInt
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "old", &oldVal, "new", &newVal); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicFloat)
	return starlark.Bool(recv.val.CAS(oldVal.GoFloat(), newVal.GoFloat())), nil
}

func floatAdd(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var delta itn.FloatOrInt
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "delta", &delta); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicFloat)
	return starlark.Float(recv.val.Add(delta.GoFloat())), nil
}

func floatSub(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var delta itn.FloatOrInt
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "delta", &delta); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicFloat)
	return starlark.Float(recv.val.Sub(delta.GoFloat())), nil
}

// for string
/*
new_string(value: string) -> AtomicString

AtomicString.get() -> string
AtomicString.set(value: string)
AtomicString.cas(old: string, new: string) -> bool
*/

var (
	stringMethods = map[string]*starlark.Builtin{
		"get": starlark.NewBuiltin("get", stringGet),
		"set": starlark.NewBuiltin("set", stringSet),
		"cas": starlark.NewBuiltin("cas", stringCAS),
	}
)

func stringGet(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicString)
	return starlark.String(recv.val.Load()), nil
}

func stringSet(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var value string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "value", &value); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicString)
	recv.val.Store(value)
	return starlark.None, nil
}

func stringCAS(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var oldVal, newVal string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "old", &oldVal, "new", &newVal); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicString)
	return starlark.Bool(recv.val.CompareAndSwap(oldVal, newVal)), nil
}
