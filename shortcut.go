package starlet

import "io/fs"

// RunScript initiates a Machine, executes a script with extra variables, and returns the Machine and the execution result.
func RunScript(content []byte, extras StringAnyMap) (*Machine, StringAnyMap, error) {
	m := NewDefault()
	res, err := m.RunScript(content, extras)
	return m, res, err
}

// RunFile initiates a Machine, executes a script from a file with extra variables, and returns the Machine and the execution result.
func RunFile(name string, fileSys fs.FS, extras StringAnyMap) (*Machine, StringAnyMap, error) {
	m := NewDefault()
	res, err := m.RunFile(name, fileSys, extras)
	return m, res, err
}

// RunTrustedScript initiates a Machine, executes a script with all builtin modules loaded and extra variables, returns the Machine and the result.
// Use with caution as it allows script access to file system and network.
func RunTrustedScript(content []byte, globals, extras StringAnyMap) (*Machine, StringAnyMap, error) {
	ab := ListBuiltinModules()
	m := NewWithNames(globals, ab, ab)
	res, err := m.RunScript(content, extras)
	return m, res, err
}

// RunTrustedFile initiates a Machine, executes a script from a file with all builtin modules loaded and extra variables, returns the Machine and the result.
// Use with caution as it allows script access to file system and network.
func RunTrustedFile(name string, fileSys fs.FS, globals, extras StringAnyMap) (*Machine, StringAnyMap, error) {
	ab := ListBuiltinModules()
	m := NewWithNames(globals, ab, ab)
	res, err := m.RunFile(name, fileSys, extras)
	return m, res, err
}
