package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"go.starlark.net/starlark"
)

// maxCachedSchemas bounds the compiled-schema cache: scripts can mint
// unbounded distinct schema texts, and compiled schemas hold their full
// structure. On overflow the cache is dropped wholesale (compilation is
// cheap relative to unbounded growth).
const maxCachedSchemas = 64

// maxValidateErrors caps how many violations are listed in a validation
// message; the remainder is summarized.
const maxValidateErrors = 10

var (
	schemaCacheMu sync.Mutex
	schemaCache   = make(map[string]*jsonschema.Schema)
)

// compileSchema compiles (with caching) a JSON Schema from its text. $ref to
// external resources — files or the network — is blocked, keeping the json
// module pure: a schema must be self-contained.
func compileSchema(text string) (*jsonschema.Schema, error) {
	schemaCacheMu.Lock()
	defer schemaCacheMu.Unlock()
	if s, ok := schemaCache[text]; ok {
		return s, nil
	}
	c := jsonschema.NewCompiler()
	c.LoadURL = func(s string) (io.ReadCloser, error) {
		return nil, fmt.Errorf("external $ref %q not allowed: schemas must be self-contained", s)
	}
	if err := c.AddResource("inline://schema", strings.NewReader(text)); err != nil {
		return nil, err
	}
	s, err := c.Compile("inline://schema")
	if err != nil {
		return nil, err
	}
	if len(schemaCache) >= maxCachedSchemas {
		schemaCache = make(map[string]*jsonschema.Schema)
	}
	schemaCache[text] = s
	return s, nil
}

// formatValidationError flattens a validation result into one line per
// violation, each prefixed with the JSON Pointer of the offending location.
func formatValidationError(ve *jsonschema.ValidationError) string {
	var lines []string
	for _, u := range ve.BasicOutput().Errors {
		if u.KeywordLocation == "" { // the root "doesn't validate with ..." wrapper
			continue
		}
		loc := u.InstanceLocation
		if loc == "" {
			loc = "/"
		}
		lines = append(lines, fmt.Sprintf("at %s: %s", loc, u.Error))
	}
	if len(lines) > maxValidateErrors {
		rest := len(lines) - maxValidateErrors
		lines = append(lines[:maxValidateErrors], fmt.Sprintf("... and %d more", rest))
	}
	return strings.Join(lines, "\n")
}

// generateValidate builds json.validate / json.try_validate, checking a JSON
// document against a JSON Schema (drafts 4/6/7/2019-09/2020-12, detected from
// $schema; default 2020-12). data and schema each accept a JSON string/bytes
// or a Starlark value.
//
// validate returns None when the data conforms and fails with a
// pointer-per-violation message otherwise. try_validate distinguishes the
// three outcomes: (True, None) valid, (False, details) invalid, and
// (None, error) when validation could not run (bad schema, bad arguments,
// malformed JSON text).
func generateValidate(try bool) func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
	return func(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var data, schema starlark.Value
		if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "data", &data, "schema", &schema); err != nil {
			return failValidate(try, err, fn)
		}
		compiled, doc, err := prepareValidation(data, schema)
		if err != nil {
			return failValidate(try, err, fn)
		}

		verr := compiled.Validate(doc)
		if verr == nil {
			if try {
				return starlark.Tuple{starlark.True, none}, nil
			}
			return none, nil
		}
		ve, ok := verr.(*jsonschema.ValidationError)
		if !ok {
			return failValidate(try, verr, fn)
		}
		details := formatValidationError(ve)
		if try {
			return starlark.Tuple{starlark.False, starlark.String(details)}, nil
		}
		return none, fmt.Errorf("%s: data does not conform to the schema:\n%s", fn.Name(), details)
	}
}

// prepareValidation resolves the schema (compiled, cached) and decodes the
// data document; any error here means validation could not run at all.
func prepareValidation(data, schema starlark.Value) (*jsonschema.Schema, interface{}, error) {
	schemaBytes, err := getJsonBytes(schema)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid schema: %w", err)
	}
	compiled, err := compileSchema(string(schemaBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("invalid schema: %w", err)
	}
	dataBytes, err := getJsonBytes(data)
	if err != nil {
		return nil, nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(dataBytes))
	dec.UseNumber()
	var doc interface{}
	if err := dec.Decode(&doc); err != nil {
		return nil, nil, fmt.Errorf("invalid data: %w", err)
	}
	// Validate a single JSON document. Trailing content (a second document
	// or garbage) means the whole input was never checked, so reject it
	// rather than silently validating only the first value. dec.More()
	// stays false for trailing whitespace.
	if dec.More() {
		return nil, nil, fmt.Errorf("invalid data: unexpected trailing content after JSON document")
	}
	return compiled, doc, nil
}

// failValidate shapes a cannot-run failure (bad schema/arguments/JSON): the
// try_ variant returns (None, message) — distinct from (False, details) for
// data that was validated and found invalid.
func failValidate(try bool, err error, fn *starlark.Builtin) (starlark.Value, error) {
	if try {
		return starlark.Tuple{none, starlark.String(err.Error())}, nil
	}
	return none, fmt.Errorf("%s: %w", fn.Name(), err)
}
