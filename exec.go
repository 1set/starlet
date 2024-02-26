package starlet

import (
	"bytes"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func starlarkExecFile(opts *syntax.FileOptions, thread *starlark.Thread, filename string, src interface{}, predeclared starlark.StringDict) (starlark.StringDict, error) {
	// Parse, resolve, and compile a Starlark source file.
	_, prog, err := starlark.SourceProgramOptions(opts, filename, src, predeclared.Has)
	if err != nil {
		return nil, err
	}

	// Try to save it to the cache
	buf := new(bytes.Buffer)
	if err := prog.Write(buf); err != nil {
		return nil, err
	}
	bs := buf.Bytes()
	//cv := starlark.CompilerVersion

	// Reload as a new program
	np, err := starlark.CompiledProgram(bytes.NewReader(bs))
	if err != nil {
		return nil, err
	}
	prog = np

	g, err := prog.Init(thread, predeclared)
	g.Freeze()
	return g, err
}
