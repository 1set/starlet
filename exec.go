package starlet

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	itn "github.com/1set/starlet/internal"
	"go.starlark.net/starlark"
)

// execStarlarkFile executes a Starlark file with the given filename and source, and returns the global environment and any error encountered.
// If the cache is enabled, it will try to load the compiled program from the cache first, and save the compiled program to the cache after compilation.
func (m *Machine) execStarlarkFile(filename string, src interface{}, allowCache bool) (starlark.StringDict, error) {
	// restore the arguments for starlark.ExecFileOptions
	opts := m.getFileOptions()
	thread := m.thread
	predeclared := m.predeclared
	hasCache := m.progCache != nil

	// if cache is not enabled or not allowed, just execute the original source
	if !hasCache || !allowCache {
		return starlark.ExecFileOptions(opts, thread, filename, src, predeclared)
	}

	// for compiled program and cache key
	var (
		prog *starlark.Program
		err  error
		//key = fmt.Sprintf("%d:%s", starlark.CompilerVersion, filename)
		key = getCacheKey(filename, src)
	)

	// try to load compiled program from cache first
	if hasCache {
		// if cache is enabled, try to load compiled bytes from cache first
		if cb, ok := m.progCache.Get(key); ok {
			// load program from compiled bytes
			if prog, err = starlark.CompiledProgram(bytes.NewReader(cb)); err != nil {
				// if failed, remove the result and continue
				prog = nil
			}
		}
	}

	// if program is not loaded from cache, compile and cache it
	if prog == nil {
		// parse, resolve, and compile a Starlark source file.
		if _, prog, err = starlark.SourceProgramOptions(opts, filename, src, predeclared.Has); err != nil {
			return nil, err
		}
		// dump the compiled program to bytes
		buf := new(bytes.Buffer)
		if err = prog.Write(buf); err != nil {
			return nil, err
		}
		// save the compiled bytes to cache
		_ = m.progCache.Set(key, buf.Bytes())
	}

	// execute the compiled program
	g, err := prog.Init(thread, predeclared)
	g.Freeze()
	return g, err
}

func getCacheKey(filename string, src interface{}) string {
	var k string
	switch s := src.(type) {
	case string:
		k = itn.GetStringMD5(s)
	case []byte:
		k = itn.GetBytesMD5(s)
	default:
		k = filename
	}
	return fmt.Sprintf("%d:%s", starlark.CompilerVersion, k)
}

// ByteCache is an interface for caching byte data, used for caching compiled Starlark programs.
type ByteCache interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte) error
}

// MemoryCache is a simple in-memory map-based ByteCache, serves as a default cache for Starlark programs.
type MemoryCache struct {
	sync.RWMutex
	data map[string][]byte
}

// NewMemoryCache creates a new MemoryCache instance.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string][]byte),
	}
}

// Get returns the value for the given key, and whether the key exists.
func (c *MemoryCache) Get(key string) ([]byte, bool) {
	c.RLock()
	defer c.RUnlock()

	if c == nil || c.data == nil {
		return nil, false
	}
	v, ok := c.data[key]
	return v, ok
}

// Set sets the value for the given key.
func (c *MemoryCache) Set(key string, value []byte) error {
	c.Lock()
	defer c.Unlock()

	if c == nil || c.data == nil {
		return errors.New("no data map found in the cache")
	}
	c.data[key] = value
	return nil
}
