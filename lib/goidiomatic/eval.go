// The expression helpers provide data-only, read-only Starlark evaluation.
// They use explicit variables, preserve their native types, and do not expose
// caller functions, modules, host wrappers, or resource handles. This is not a
// process sandbox or a hard allocation limit; the caller owns execution limits.
package goidiomatic

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"unicode"
	"unicode/utf8"

	"go.starlark.net/resolve"
	std "go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// Error identifies the stage of a failed expression evaluation. Cause retains
// parser/resolver errors and EvalError backtraces for errors.As/errors.Is.
type Error struct {
	Phase string
	Kind  string
	Cause error
}

func (e *Error) Error() string {
	message, _ := diagnosticText(e.Cause.Error(), maxErrorMessage)
	return "eval: " + e.Phase + ": " + message
}
func (e *Error) Unwrap() error { return e.Cause }

func evalError(phase string, cause error) *Error {
	return &Error{Phase: phase, Kind: phase, Cause: cause}
}

var (
	// Builtins are immutable and capture only their fixed error-handling mode.
	// Reuse them across fresh module dictionaries; no Thread or vars are cached.
	evalBuiltin    = std.NewBuiltin("eval", evaluationBuiltin(false))
	tryEvalBuiltin = std.NewBuiltin("try_eval", evaluationBuiltin(true))
	// Capture the standard, immutable builtin instances before evaluation.
	// Like upstream, this package requires Universe not be mutated concurrently
	// with execution. A later replacement of an allowed name is not trusted.
	pureBuiltins = captureBuiltins()
	probeOnce    sync.Once
	threadProbe  *std.Function
)

func captureBuiltins() std.StringDict {
	names := []string{"None", "True", "False", "abs", "any", "all", "bool", "bytes", "chr", "dict", "dir", "enumerate", "fail", "float", "getattr", "hasattr", "hash", "int", "len", "list", "max", "min", "ord", "range", "repr", "reversed", "set", "sorted", "str", "tuple", "type", "zip"}
	result := make(std.StringDict, len(names))
	for _, name := range names {
		result[name] = std.Universe[name]
	}
	return result
}

func makeThreadProbe() *std.Function {
	f, err := std.ExprFuncOptions(&syntax.FileOptions{Set: true}, "<eval.check>", "0", nil)
	if err != nil {
		panic(err) // A fixed integer literal cannot fail to compile.
	}
	f.Freeze()
	return f
}

// checkThread runs a constant, closed expression on the caller's existing VM.
// This checks its private cancellation/step state without inspecting internals,
// resetting limits, or relying on the wording of an EvalError. In particular,
// an OnMaxSteps panic propagates unchanged to the caller's execution boundary.
func checkThread(thread *std.Thread) error {
	if thread == nil {
		return evalError("input", fmt.Errorf("thread is nil"))
	}
	if ctx, ok := thread.Local("context").(context.Context); ok && ctx != nil {
		if err := ctx.Err(); err != nil {
			return evalError("execution", err)
		}
	}
	probeOnce.Do(func() { threadProbe = makeThreadProbe() })
	if _, err := std.Call(thread, threadProbe, nil, nil); err != nil {
		return evalError("execution", err)
	}
	return nil
}

// Eval evaluates one Starlark expression against an owned, frozen copy of vars.
// The caller must provide an active or exclusively owned Thread and either an
// exclusively owned/synchronized variables graph or already-frozen native
// data. Merely avoiding writes does not make an unfrozen set safe to share:
// native iteration updates its bookkeeping. Values and names are checked
// before interpretation; results are copied, frozen native data only.
//
// Eval preserves the caller's cancellation and cumulative step accounting.
// It does not set a new step limit, recover caller panics, or capture locals.
func Eval(thread *std.Thread, source string, vars std.StringDict) (std.Value, error) {
	if err := checkThread(thread); err != nil {
		return std.None, err
	}
	expr, phase, err := parseSource(source)
	if err != nil {
		return std.None, evalError(phase, err)
	}
	if len(vars) > maxDataNodes {
		return std.None, evalError("input", fmt.Errorf("too many variables"))
	}
	nameBytes := 0
	for name := range vars {
		if len(name) > maxDataBytes-nameBytes {
			return std.None, evalError("input", fmt.Errorf("variable names exceed data byte limit"))
		}
		nameBytes += len(name)
	}
	copy := newDataCopy(thread)
	env := make(std.StringDict, len(vars)+len(std.Universe))
	names := vars.Keys()
	for _, name := range names {
		if err := copy.chargeBytes(len(name), "vars"); err != nil {
			return std.None, evalError("input", err)
		}
		if !validName(name) {
			preview, _ := diagnosticText(name, 64)
			return std.None, evalError("input", fmt.Errorf("invalid variable name %q (%d bytes)", preview, len(name)))
		}
		value, err := copy.value(vars[name], "vars."+name, 0)
		if err != nil {
			return std.None, wrapInputError(err)
		}
		env[name] = value
	}
	// Only private copies are frozen, after the full input graph was accepted.
	env.Freeze()
	for name := range std.Universe {
		if _, exists := env[name]; exists {
			continue
		}
		if value, approved := pureBuiltins[name]; approved {
			env[name] = value
		} else {
			env[name] = deniedBuiltin(name)
		}
	}
	// Approved instances remain usable even if another host replaced or removed
	// a Universe entry before execution. Data variables keep normal shadowing.
	for name, value := range pureBuiltins {
		if _, exists := env[name]; !exists {
			env[name] = value
		}
	}
	if err := checkThread(thread); err != nil {
		return std.None, err
	}
	value, err := std.EvalExprOptions(&syntax.FileOptions{Set: true}, thread, expr, env)
	if err != nil {
		if terminal := checkThread(thread); terminal != nil {
			return std.None, terminal
		}
		var own *Error
		if errors.As(err, &own) {
			return std.None, &Error{Phase: own.Phase, Kind: own.Kind, Cause: err}
		}
		var unresolved resolve.ErrorList
		if errors.As(err, &unresolved) {
			return std.None, evalError("resolve", err)
		}
		return std.None, evalError("runtime", err)
	}
	result, err := newDataCopy(thread).value(value, "result", 0)
	if err != nil {
		var terminal *Error
		if errors.As(err, &terminal) {
			return std.None, err
		}
		return std.None, evalError("output", err)
	}
	if err := checkThread(thread); err != nil {
		return std.None, err
	}
	result.Freeze()
	return result, nil
}

func wrapInputError(err error) error {
	var execution *Error
	if errors.As(err, &execution) {
		return err
	}
	return evalError("input", err)
}

func validName(name string) bool {
	if len(name) == 0 || len(name) > 512 || !utf8.ValidString(name) {
		return false
	}
	for i, r := range name {
		if r != '_' && !unicode.IsLetter(r) && (i == 0 || !unicode.IsDigit(r)) {
			return false
		}
	}
	expr, err := (&syntax.FileOptions{}).ParseExpr("<eval.vars>", name, 0)
	if err != nil {
		return false
	}
	ident, ok := expr.(*syntax.Ident)
	return ok && ident.Name == name
}

func deniedBuiltin(name string) *std.Builtin {
	return std.NewBuiltin(name, func(*std.Thread, *std.Builtin, std.Tuple, []std.Tuple) (std.Value, error) {
		return std.None, &Error{Phase: "policy", Kind: "capability", Cause: fmt.Errorf("builtin %q is unavailable", name)}
	})
}

func evaluationBuiltin(try bool) func(*std.Thread, *std.Builtin, std.Tuple, []std.Tuple) (std.Value, error) {
	return func(thread *std.Thread, fn *std.Builtin, args std.Tuple, kwargs []std.Tuple) (std.Value, error) {
		if err := checkThread(thread); err != nil {
			return std.None, err
		}
		var sourceValue std.Value
		var varsValue std.Value
		err := std.UnpackArgs(fn.Name(), args, kwargs, "source", &sourceValue, "vars?", &varsValue)
		vars := make(std.StringDict)
		source, sourceOK := sourceValue.(std.String)
		if err != nil {
			err = evalError("arguments", err)
		} else if !sourceOK {
			err = evalError("input", fmt.Errorf("source must be a native string, got %T", sourceValue))
		} else if varsValue != nil {
			dict, ok := varsValue.(*std.Dict)
			if !ok || dict == nil {
				err = evalError("input", fmt.Errorf("vars must be a native dict"))
			} else if dict.Len() > maxDataNodes {
				err = evalError("input", fmt.Errorf("too many variables"))
			} else {
				for _, pair := range dict.Items() {
					key, ok := pair[0].(std.String)
					if !ok {
						err = evalError("input", fmt.Errorf("vars keys must be strings"))
						break
					}
					vars[string(key)] = pair[1]
				}
			}
		}
		value := std.Value(std.None)
		if err == nil {
			value, err = Eval(thread, string(source), vars)
		}
		if terminal := checkThread(thread); terminal != nil {
			return std.None, terminal
		}
		if err != nil {
			var execution *Error
			if errors.As(err, &execution) && execution.Phase == "execution" {
				// Preserve a terminal failure even if the owner concurrently
				// cleared the Thread cancellation before the final probe.
				return std.None, err
			}
			if try {
				return std.Tuple{std.None, errorInfo(err)}, nil
			}
			return std.None, err
		}
		if try {
			return std.Tuple{value, std.None}, nil
		}
		return value, nil
	}
}

const (
	maxErrorMessage = 4096
	maxErrorFrames  = 32
)

func diagnosticText(text string, limit int) (string, bool) {
	if len(text) <= limit {
		return text, false
	}
	const suffix = "... (truncated)"
	end := limit - len(suffix)
	for end > 0 && !utf8.RuneStart(text[end]) {
		end--
	}
	return text[:end] + suffix, true
}

func errorInfo(err error) *std.Dict {
	phase, kind := "runtime", "runtime"
	var own *Error
	if errors.As(err, &own) {
		phase, kind = own.Phase, own.Kind
	}
	position := std.Value(std.None)
	frames := std.Tuple{}
	message, truncated := diagnosticText(err.Error(), maxErrorMessage)
	var parse syntax.Error
	var unresolved resolve.ErrorList
	var runtime *std.EvalError
	if errors.As(err, &parse) {
		position = positionInfo(parse.Pos)
		truncated = truncated || len(parse.Pos.Filename()) > 512
	} else if errors.As(err, &unresolved) && len(unresolved) > 0 {
		position = positionInfo(unresolved[0].Pos)
		truncated = truncated || len(unresolved[0].Pos.Filename()) > 512
	} else if errors.As(err, &runtime) {
		start := 0
		if len(runtime.CallStack) > maxErrorFrames {
			start = len(runtime.CallStack) - maxErrorFrames
			truncated = true
		}
		for _, frame := range runtime.CallStack[start:] {
			f := positionInfo(frame.Pos)
			truncated = truncated || len(frame.Pos.Filename()) > 512
			name, shortened := diagnosticText(frame.Name, 256)
			truncated = truncated || shortened
			_ = f.SetKey(std.String("name"), std.String(name))
			frames = append(frames, f)
			if frame.Pos.IsValid() && frame.Pos.Filename() != "<builtin>" {
				position = positionInfo(frame.Pos)
			}
		}
	}
	info := std.NewDict(6)
	for _, pair := range []std.Tuple{{std.String("phase"), std.String(phase)}, {std.String("kind"), std.String(kind)}, {std.String("message"), std.String(message)}, {std.String("position"), position}, {std.String("frames"), frames}, {std.String("truncated"), std.Bool(truncated)}} {
		_ = info.SetKey(pair[0], pair[1])
	}
	info.Freeze()
	return info
}

func positionInfo(pos syntax.Position) *std.Dict {
	d := std.NewDict(3)
	file, _ := diagnosticText(pos.Filename(), 512)
	_ = d.SetKey(std.String("file"), std.String(file))
	_ = d.SetKey(std.String("line"), std.MakeInt(int(pos.Line)))
	_ = d.SetKey(std.String("column"), std.MakeInt(int(pos.Col)))
	return d
}
