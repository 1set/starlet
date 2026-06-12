// Package path provides path manipulation for Starlark. The lexical
// functions (join, basename, dirname, normpath, split, splitext, isabs,
// relpath) follow Python's posixpath semantics - pure string work on
// "/"-separated paths, identical on every OS - while the filesystem
// functions (abs, exists, is_file, is_dir, is_link, listdir, getcwd,
// chdir, mkdir) operate on the real, OS-native filesystem.
//
// WARNING: chdir changes the working directory of the WHOLE PROCESS - it
// affects every machine and goroutine in the host, not just the calling
// script, and concurrent machines calling it race with each other. Never
// expose this module to untrusted scripts.
package path

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	tps "github.com/1set/starlet/dataconv/types"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used in starlark's load() function, eg: load('path', 'join')
const ModuleName = "path"

var (
	once       sync.Once
	pathModule starlark.StringDict
)

// LoadModule loads the path module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		pathModule = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"abs":     starlark.NewBuiltin(ModuleName+".abs", absPath),
					"try_abs": starlark.NewBuiltin(ModuleName+".try_abs", wrapTry(absPath)),
					"join":    starlark.NewBuiltin(ModuleName+".join", joinPaths),
					"exists":  wrapExistPath("exists", checkExistPath),
					"is_file": wrapExistPath("is_file", checkFileExist),
					"is_dir":  wrapExistPath("is_dir", checkDirExist),
					"is_link": wrapExistPath("is_link", checkSymlinkExist),
					"listdir": starlark.NewBuiltin(ModuleName+".listdir", listDirContents),
					"getcwd":  starlark.NewBuiltin(ModuleName+".getcwd", getCWD),
					"chdir":   starlark.NewBuiltin(ModuleName+".chdir", changeCWD),
					"mkdir":   starlark.NewBuiltin(ModuleName+".mkdir", makeDir),
					// lexical functions with Python (posixpath) semantics:
					// pure string work on "/"-separated paths, no cleaning
					// beyond what each function defines, identical on every OS
					"basename":    genLexical1("basename", pyBasename),
					"dirname":     genLexical1("dirname", pyDirname),
					"normpath":    genLexical1("normpath", pyNormpath),
					"expanduser":  starlark.NewBuiltin(ModuleName+".expanduser", expandUser),
					"isabs":       starlark.NewBuiltin(ModuleName+".isabs", isAbs),
					"split":       starlark.NewBuiltin(ModuleName+".split", splitPath),
					"splitext":    starlark.NewBuiltin(ModuleName+".splitext", splitExt),
					"relpath":     starlark.NewBuiltin(ModuleName+".relpath", relPath),
					"try_relpath": starlark.NewBuiltin(ModuleName+".try_relpath", wrapTry(relPath)),
				},
			},
		}
	})
	return pathModule, nil
}

// absPath returns the absolute representation of path.
func absPath(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path); err != nil {
		return nil, err
	}
	// get absolute path
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return starlark.String(abs), nil
}

// joinPaths joins any number of path elements into a single path.
func joinPaths(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// check arguments
	if len(args) < 1 {
		return nil, fmt.Errorf("%s: got %d arguments, want at least 1", b.Name(), len(args))
	}
	// unpack arguments
	paths := make([]string, len(args))
	for i, arg := range args {
		s, ok := starlark.AsString(arg)
		if !ok {
			return nil, fmt.Errorf("%s: for parameter path: got %s, want string", b.Name(), arg.Type())
		}
		paths[i] = s
	}
	// join paths with Python (posixpath) semantics: pure string joining,
	// no cleaning. An absolute component resets the result; an empty
	// component contributes a trailing separator. The previous
	// filepath.Join behavior (collapse "..", drop empty parts, ignore
	// absolute restarts, OS separator) is gone - see the README migration
	// notes for the differences.
	joined := paths[0]
	for _, p := range paths[1:] {
		switch {
		case strings.HasPrefix(p, "/"):
			joined = p
		case joined == "" || strings.HasSuffix(joined, "/"):
			joined += p
		default:
			joined += "/" + p
		}
	}
	return starlark.String(joined), nil
}

// wrapTry converts a builtin into its try_ variant: instead of aborting the
// whole script on failure, it returns a (value, error-string) pair with the
// Go error always nil - the shape shared with the json/csv/http modules.
func wrapTry(fn func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		res, err := fn(thread, b, args, kwargs)
		if err != nil {
			return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
		}
		if res == nil {
			res = starlark.None
		}
		return starlark.Tuple{res, starlark.None}, nil
	}
}

// genLexical1 wraps a pure string->string lexical helper as a builtin.
func genLexical1(name string, fn func(string) string) *starlark.Builtin {
	return starlark.NewBuiltin(ModuleName+"."+name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var path string
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path); err != nil {
			return nil, err
		}
		return starlark.String(fn(path)), nil
	})
}

// The helpers below hand-implement Python's posixpath semantics. The Go
// standard library cannot be delegated to: path.Join/Clean normalize
// eagerly, filepath.Base("a/b/") is "b" while Python's basename is "",
// filepath.Dir cleans its result, and so on - each function is matched
// against CPython's posixpath behavior in the tests.

// pyBasename returns the final component of a path: everything after the
// last slash, so a path with a trailing slash has an empty basename.
func pyBasename(p string) string {
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		return p[i+1:]
	}
	return p
}

// pyDirname returns the directory part of a path: everything before the
// last slash, with trailing slashes stripped unless the result is all
// slashes (so dirname("//a") keeps "//").
func pyDirname(p string) string {
	i := strings.LastIndexByte(p, '/') + 1
	head := p[:i]
	if head != "" && head != strings.Repeat("/", len(head)) {
		head = strings.TrimRight(head, "/")
	}
	return head
}

// pyNormpath collapses redundant separators and up-level references
// lexically: "a//b", "a/./b" and "a/c/../b" all become "a/b". As in
// Python, exactly two leading slashes are preserved (POSIX gives them
// implementation-defined meaning) and an empty path normalizes to ".".
func pyNormpath(p string) string {
	if p == "" {
		return "."
	}
	initialSlashes := 0
	if strings.HasPrefix(p, "/") {
		initialSlashes = 1
		if strings.HasPrefix(p, "//") && !strings.HasPrefix(p, "///") {
			initialSlashes = 2
		}
	}
	var comps []string
	for _, comp := range strings.Split(p, "/") {
		if comp == "" || comp == "." {
			continue
		}
		if comp != ".." || (initialSlashes == 0 && len(comps) == 0) || (len(comps) > 0 && comps[len(comps)-1] == "..") {
			comps = append(comps, comp)
		} else if len(comps) > 0 {
			comps = comps[:len(comps)-1]
		}
	}
	res := strings.Repeat("/", initialSlashes) + strings.Join(comps, "/")
	if res == "" {
		return "."
	}
	return res
}

// expandUser replaces a leading "~" with the current user's home
// directory. Only the bare "~" form expands; "~user" is returned
// unchanged (matching Python when the user lookup is unavailable), and so
// is the path when the home directory cannot be determined.
func expandUser(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path); err != nil {
		return nil, err
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return starlark.String(home + path[1:]), nil
		}
	}
	return starlark.String(path), nil
}

// isAbs reports whether a path is absolute in the posix sense (starts
// with a slash).
func isAbs(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path); err != nil {
		return nil, err
	}
	return starlark.Bool(strings.HasPrefix(path, "/")), nil
}

// splitPath splits a path into a (head, tail) pair: tail is everything
// after the last slash, head is the rest with trailing slashes stripped
// unless it is all slashes - so split("/a") is ("/", "a") and
// split("a/b/") is ("a/b", "").
func splitPath(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path); err != nil {
		return nil, err
	}
	return starlark.Tuple{starlark.String(pyDirname(path)), starlark.String(pyBasename(path))}, nil
}

// splitExt splits a path into a (root, extension) pair. The extension is
// the suffix beginning at the last dot in the final component; leading
// dots do not count, so splitext(".bashrc") is (".bashrc", "").
func splitExt(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path); err != nil {
		return nil, err
	}
	sepIdx := strings.LastIndexByte(path, '/')
	dotIdx := strings.LastIndexByte(path, '.')
	if dotIdx > sepIdx {
		// the dot must come after at least one non-dot character of the
		// final component, otherwise the name is dot-prefixed, not suffixed
		for i := sepIdx + 1; i < dotIdx; i++ {
			if path[i] != '.' {
				return starlark.Tuple{starlark.String(path[:dotIdx]), starlark.String(path[dotIdx:])}, nil
			}
		}
	}
	return starlark.Tuple{starlark.String(path), starlark.String("")}, nil
}

// relPath returns a relative path to path from the start directory,
// computed lexically on the normalized inputs. Both paths must be of the
// same kind: mixing an absolute path with a relative one is an error here,
// because resolving the mix would silently depend on the process working
// directory (Python's os.path.relpath does exactly that; use abs() first
// when that behavior is wanted).
func relPath(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path, start string
	start = "."
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path, "start?", &start); err != nil {
		return nil, err
	}
	if path == "" {
		return nil, fmt.Errorf("%s: no path specified", b.Name())
	}
	if start == "" {
		start = "."
	}
	if strings.HasPrefix(path, "/") != strings.HasPrefix(start, "/") {
		return nil, fmt.Errorf("%s: cannot mix an absolute path with a relative start (resolve with abs() first)", b.Name())
	}
	splitComps := func(p string) []string {
		var out []string
		for _, c := range strings.Split(pyNormpath(p), "/") {
			if c != "" && c != "." {
				out = append(out, c)
			}
		}
		return out
	}
	pc, sc := splitComps(path), splitComps(start)
	i := 0
	for i < len(pc) && i < len(sc) && pc[i] == sc[i] {
		i++
	}
	var comps []string
	for j := i; j < len(sc); j++ {
		comps = append(comps, "..")
	}
	comps = append(comps, pc[i:]...)
	if len(comps) == 0 {
		return starlark.String("."), nil
	}
	return starlark.String(strings.Join(comps, "/")), nil
}

// wrapExistPath wraps the existPath function to be used in Starlark with a given function to check if the path exists.
func wrapExistPath(funcName string, workLoad func(path string) bool) starlark.Callable {
	return starlark.NewBuiltin(ModuleName+"."+funcName, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var path string
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path); err != nil {
			return starlark.None, err
		}
		return starlark.Bool(workLoad(path)), nil
	})
}

// checkExistPath returns true if the path exists, if it's a symbolic link, the symbolic link is followed.
func checkExistPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// checkFileExist returns true if the file exists, if it's a symbolic link, the symbolic link is followed.
func checkFileExist(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info != nil && info.Mode().IsRegular()
}

// checkDirExist returns true if the directory exists, if it's a symbolic link, the symbolic link is followed.
func checkDirExist(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info != nil && info.IsDir()
}

// checkSymlinkExist returns true if the symbolic link exists.
func checkSymlinkExist(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info != nil && info.Mode()&os.ModeSymlink == os.ModeSymlink
}

// listDirContents returns a list of directory contents.
func listDirContents(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		path       string
		recursive  bool
		filterFunc = tps.NullableCallable{}
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path, "recursive?", &recursive, "filter?", &filterFunc); err != nil {
		return nil, err
	}
	// get filter func
	var ff starlark.Callable
	if !filterFunc.IsNull() {
		ff = filterFunc.Value()
	}

	// check root stat
	rootInfo, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", b.Name(), err)
	}

	// check if path is a directory, if not return empty list
	var sl []starlark.Value
	if !rootInfo.IsDir() {
		return starlark.NewList(sl), nil
	}

	// scan directory contents
	if err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// skip same path to avoid infinite loop in case of symbolic links
		if os.SameFile(rootInfo, info) {
			return nil
		}
		// filter path
		sp := starlark.String(p)
		if ff != nil {
			filtered, err := starlark.Call(thread, ff, starlark.Tuple{sp}, nil)
			if err != nil {
				return fmt.Errorf("filter %q: %w", p, err)
			}
			if fb, ok := filtered.(starlark.Bool); !ok {
				return fmt.Errorf("filter %q: got %s, want bool", p, filtered.Type())
			} else if fb == false {
				return nil // skip path
			}
		}

		// add path to list
		sl = append(sl, sp)

		// check if we should list recursively
		if !recursive && p != path && info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("%s: %w", b.Name(), err)
	}
	return starlark.NewList(sl), nil
}

// getCWD returns the current working directory.
func getCWD(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// check the arguments: no arguments
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	// get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", b.Name(), err)
	}
	return starlark.String(cwd), nil
}

// changeCWD changes the current working directory of the whole process:
// the effect is global, leaks across machines and threads, and persists
// after the script ends. See the package warning..
func changeCWD(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path); err != nil {
		return nil, err
	}
	// change working directory
	if err := os.Chdir(path); err != nil {
		return nil, fmt.Errorf("%s: %w", b.Name(), err)
	}
	return starlark.None, nil
}

// makeDir creates a directory with the given name.
func makeDir(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		pathVal tps.StringOrBytes
		modeVal = uint32(0755)
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &pathVal, "mode?", &modeVal); err != nil {
		return starlark.None, err
	}
	// do the work
	mode := os.FileMode(modeVal)
	return starlark.None, os.MkdirAll(pathVal.GoString(), mode)
}
