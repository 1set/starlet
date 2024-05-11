package starlet

import "go.starlark.net/starlark"

// StringAnyMap type is a map of string to interface{} and is used to store global variables like StringDict of Starlark, but not a Starlark type.
type StringAnyMap map[string]interface{}

// Clone returns a copy of the data store. It returns an empty map if the current data store is nil.
func (d StringAnyMap) Clone() StringAnyMap {
	clone := make(StringAnyMap)
	for k, v := range d {
		clone[k] = v
	}
	return clone
}

// Merge merges the given data store into the current data store. It does nothing if the current data store is nil.
func (d StringAnyMap) Merge(other StringAnyMap) {
	if d == nil {
		return
	}
	for k, v := range other {
		d[k] = v
	}
}

// MergeDict merges the given string dict into the current data store.
func (d StringAnyMap) MergeDict(other starlark.StringDict) {
	if d == nil {
		return
	}
	for k, v := range other {
		d[k] = v
	}
}
