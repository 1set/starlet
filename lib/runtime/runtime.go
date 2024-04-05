// Package runtime implements the Starlark module for Go and app runtime information.
package runtime

import (
	"os"
	grt "runtime"
	"sync"
	"time"

	"github.com/1set/starlet/dataconv"
	stdtime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used in starlark's load() function, eg: load('base64', 'encode')
const ModuleName = "runtime"

var (
	once       sync.Once
	moduleData starlark.StringDict
)

// LoadModule loads the runtime module. It is concurrency-safe and idempotent.
func LoadModule() (md starlark.StringDict, err error) {
	once.Do(func() {
		var host, pwd string
		if host, err = os.Hostname(); err != nil {
			return
		}
		if pwd, err = os.Getwd(); err != nil {
			return
		}
		moduleData = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"hostname":  starlark.String(host),
					"workdir":   starlark.String(pwd),
					"os":        starlark.String(grt.GOOS),
					"arch":      starlark.String(grt.GOARCH),
					"gover":     starlark.String(grt.Version()),
					"pid":       starlark.MakeInt(os.Getpid()),
					"ppid":      starlark.MakeInt(os.Getppid()),
					"uid":       starlark.MakeInt(os.Getuid()),
					"gid":       starlark.MakeInt(os.Getgid()),
					"app_start": stdtime.Time(appStart),
					"uptime":    starlark.NewBuiltin(ModuleName+".uptime", getUpTime),
					"getenv":    starlark.NewBuiltin(ModuleName+".getenv", getenv),
					"putenv":    starlark.NewBuiltin(ModuleName+".putenv", putenv),
					"setenv":    starlark.NewBuiltin(ModuleName+".setenv", putenv), // alias "setenv" to "putenv"
					"unsetenv":  starlark.NewBuiltin(ModuleName+".unsetenv", unsetenv),
				},
			},
		}
	})
	return moduleData, err
}

var (
	appStart = time.Now()
)

// getUpTime returns time elapsed since the app started.
func getUpTime(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	return stdtime.Duration(time.Since(appStart)), nil
}

// getenv returns the value of the environment variable key as a string if it exists, or default if it doesn't.
func getenv(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		key    string
		defVal starlark.Value = starlark.None
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "key", &key, "default?", &defVal); err != nil {
		return nil, err
	}
	// get the value
	if val, ok := os.LookupEnv(key); ok {
		return starlark.String(val), nil
	}
	return defVal, nil
}

// putenv sets the value of the environment variable named by the key, returning an error if any.
// value should be a string, or it will be converted to a string.
func putenv(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		key string
		val starlark.Value
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "key", &key, "value", &val); err != nil {
		return nil, err
	}
	// set the value
	err := os.Setenv(key, dataconv.StarString(val))
	return starlark.None, err
}

// unsetenv unsets a single environment variable.
func unsetenv(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "key", &key); err != nil {
		return nil, err
	}
	// unset the value
	err := os.Unsetenv(key)
	return starlark.None, err
}
