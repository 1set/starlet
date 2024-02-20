// Package runtime implements the Starlark module for Go and app runtime information.
package runtime

import (
	"os"
	grt "runtime"
	"sync"
	"time"

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
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		host, _ := os.Hostname()
		pwd, _ := os.Getwd()
		moduleData = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"hostname":  starlark.String(host),
					"workdir":   starlark.String(pwd),
					"os":        starlark.String(grt.GOOS),
					"arch":      starlark.String(grt.GOARCH),
					"pid":       starlark.MakeInt(os.Getpid()),
					"app_start": stdtime.Time(appStart),
					"uptime":    starlark.NewBuiltin(ModuleName+".uptime", getUpTime),
				},
			},
		}
	})
	return moduleData, nil
}

var (
	appStart = time.Now()
)

// getUpTime returns time elapsed since the app started.
func getUpTime(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return stdtime.Duration(time.Since(appStart)), nil
}
