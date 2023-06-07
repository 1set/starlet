package starlet

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"sync"

	"github.com/1set/starlight/convert"
	"go.starlark.net/starlark"
)

// WithGlobals returns a new Starlight cache that passes the listed global
// values to scripts loaded with the load() script function.  Note that these
// globals will *not* be passed to individual scripts you run unless you
// explicitly pass them in the Run call.
func WithGlobals(globals map[string]interface{}, fileSys fs.FS) (*Cache, error) {
	if fileSys == nil {
		return nil, fmt.Errorf("no file system given")
	}
	g, err := convert.MakeStringDict(globals)
	if err != nil {
		return nil, err
	}
	return newCache(fileSys, g), nil
}

// LoadFunc is a function that tells Starlark how to find and load other scripts
// using the load() function. If you don't use load() in your scripts, you can pass in nil.
type LoadFunc func(thread *starlark.Thread, module string) (starlark.StringDict, error)

// Cache is a cache of scripts to avoid re-reading files and reparsing them.
type Cache struct {
	mu      sync.Mutex
	cache   *cache
	fs      fs.FS
	scripts map[string]*starlark.Program
}

func newCache(fileSys fs.FS, globals starlark.StringDict) *Cache {
	c := &Cache{
		fs:      fileSys,
		scripts: map[string]*starlark.Program{},
	}
	c.cache = &cache{
		cache:    make(map[string]*entry),
		readFile: c.readFile,
		globals:  globals,
	}
	return c
}

func runLight(p *starlark.Program, globals map[string]interface{}, load LoadFunc) (map[string]interface{}, error) {
	g, err := convert.MakeStringDict(globals)
	if err != nil {
		return nil, err
	}
	ret, err := p.Init(&starlark.Thread{Load: load}, g)
	if err != nil {
		return nil, err
	}
	return convert.FromStringDict(ret), nil
}

// Run looks for a file with the given filename, and runs it with the given globals
// passed to the script's global namespace. The return value is all convertible
// global variables from the script, which may include the passed-in globals.
func (c *Cache) Run(filename string, globals map[string]interface{}) (map[string]interface{}, error) {
	dict, err := convert.MakeStringDict(globals)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	if p, ok := c.scripts[filename]; ok {
		c.mu.Unlock()
		return runLight(p, globals, c.Load)
	}
	c.mu.Unlock()

	b, err := c.readFile(filename)
	if err != nil {
		return nil, err
	}
	_, p, err := starlark.SourceProgram(filename, b, dict.Has)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.scripts[filename] = p
	c.mu.Unlock()
	return runLight(p, globals, c.Load)
}

// Load loads a module using the cache's configured directories.
func (c *Cache) Load(_ *starlark.Thread, module string) (starlark.StringDict, error) {
	// TODO: add a way to load modules from structs.
	return c.cache.Load(module)
}

// readFile reads the given filename from the given file system.
func (c *Cache) readFile(filename string) ([]byte, error) {
	if c.fs == nil {
		return nil, fmt.Errorf("no file system given")
	}
	rd, err := c.fs.Open(filename)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(rd)
}

// Reset clears all cached scripts.
func (c *Cache) Reset() {
	c.mu.Lock()
	c.scripts = map[string]*starlark.Program{}
	c.cache.reset()
	c.mu.Unlock()
}

// Forget clears the cached script for the given filename.
func (c *Cache) Forget(filename string) {
	c.mu.Lock()
	c.cache.remove(filename)
	delete(c.scripts, filename)
	c.mu.Unlock()
}
