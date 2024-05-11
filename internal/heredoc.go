package internal

import (
	"github.com/h2so5/here"
)

var (
	// HereDoc returns un-indented string as here-document.
	HereDoc = here.Doc

	// HereDocf returns unindented and formatted string as here-document. Formatting is done as for fmt.Printf().
	HereDocf = here.Docf
)
