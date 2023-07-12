// Package internal contains types and utilities that are not part of the public API, and may change without notice.
// It should be only imported by the custom Starlark modules under starlet/lib folders, and not by the Starlet main package to avoid cyclic import.
package internal

// StarletVersion is the current version of Starlet.
const StarletVersion = "v0.0.3"
