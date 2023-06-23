package starlet

import "io/fs"

// RunScript creates a new Machine, runs a script with additional variables, returns the machine and the result.
func RunScript(content []byte, extras StringAnyMap) (*Machine, StringAnyMap, error) {
	m := NewDefault()
	res, err := m.RunScript(content, extras)
	return m, res, err
}

// RunFile creates a new Machine, runs a script from a file with additional variables, returns the machine and the result.
func RunFile(name string, fileSys fs.FS, extras StringAnyMap) (*Machine, StringAnyMap, error) {
	m := NewDefault()
	res, err := m.RunFile(name, fileSys, extras)
	return m, res, err
}

// RunTrustedScript creates a new Machine, runs a script with all builtin modules loaded and variables, which is unsafe, returns the machine and the result.
// Warning: Loading all builtin modules can give the script access to the file system and network, making it unsafe.
func RunTrustedScript(content []byte, globals, extras StringAnyMap) (*Machine, StringAnyMap, error) {
	ab := ListBuiltinModules()
	m := NewWithNames(globals, ab, ab)
	res, err := m.RunScript(content, extras)
	return m, res, err
}

// RunTrustedFile creates a new Machine, runs a script from a file with all builtin modules loaded and variables, which is unsafe, returns the machine and the result.
// Warning: Loading all builtin modules can give the script access to the file system and network, making it unsafe.
func RunTrustedFile(name string, fileSys fs.FS, globals, extras StringAnyMap) (*Machine, StringAnyMap, error) {
	ab := ListBuiltinModules()
	m := NewWithNames(globals, ab, ab)
	res, err := m.RunFile(name, fileSys, extras)
	return m, res, err
}
