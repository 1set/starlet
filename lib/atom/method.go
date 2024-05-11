package atom

import (
	tps "github.com/1set/starlet/dataconv/types"
	"go.starlark.net/starlark"
)

// for integer

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
	var value tps.FloatOrInt
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "value", &value); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicFloat)
	recv.val.Store(value.GoFloat())
	return starlark.None, nil
}

func floatCAS(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var oldVal, newVal tps.FloatOrInt
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "old", &oldVal, "new", &newVal); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicFloat)
	return starlark.Bool(recv.val.CAS(oldVal.GoFloat(), newVal.GoFloat())), nil
}

func floatAdd(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var delta tps.FloatOrInt
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "delta", &delta); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicFloat)
	return starlark.Float(recv.val.Add(delta.GoFloat())), nil
}

func floatSub(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var delta tps.FloatOrInt
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "delta", &delta); err != nil {
		return nil, err
	}
	recv := b.Receiver().(*AtomicFloat)
	return starlark.Float(recv.val.Sub(delta.GoFloat())), nil
}

// for string

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
