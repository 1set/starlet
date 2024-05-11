// Package internal contains types and utilities that are not part of the public API, and may change without notice.
// It should be only imported by the custom Starlark modules under starlet/lib folders, and not by the Starlet main package to avoid cyclic import.
package internal

import "github.com/h2so5/here"

// StarletVersion should be the current version of Starlet.
const StarletVersion = "v0.1.0"

var (
	// HereDoc returns un-indented string as here-document.
	HereDoc = here.Doc

	// HereDocf returns unindented and formatted string as here-document. Formatting is done as for fmt.Printf().
	HereDocf = here.Docf
)
