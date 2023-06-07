package starlet

import "go.starlark.net/starlark"

// DataStore is a map of string to interface{}.
type DataStore map[string]interface{}

// Clone returns a copy of the data store.
func (d DataStore) Clone() DataStore {
	clone := make(DataStore)
	for k, v := range d {
		clone[k] = v
	}
	return clone
}

// Merge merges the given data store into the current data store.
func (d DataStore) Merge(other DataStore) {
	if d == nil {
		return
	}
	for k, v := range other {
		d[k] = v
	}
}

// MergeDict merges the given string dict into the current data store.
func (d DataStore) MergeDict(other starlark.StringDict) {
	if d == nil {
		return
	}
	for k, v := range other {
		d[k] = v
	}
}
