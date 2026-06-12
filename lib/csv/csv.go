// Package csv reads comma-separated values from strings and writes CSV data to strings.
//
// Migrated from https://github.com/qri-io/starlib/tree/master/encoding/csv
package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/1set/starlet/dataconv"
	tps "github.com/1set/starlet/dataconv/types"
	"github.com/1set/starlet/internal/replacecr"
	"github.com/1set/starlet/lib/file"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// cellToString renders a single cell value for CSV output. Every supported
// type is rendered explicitly: the previous fmt.Sprintf("%v", v) silently
// produced Go-flavored text — scientific notation for large/small floats
// ("1e+06"), "<nil>" for None, Go syntax for nested collections — which
// corrupted the written file without any error.
func cellToString(v interface{}) (string, error) {
	switch c := v.(type) {
	case nil:
		// matches the empty cell written for a missing write_dict key
		return "", nil
	case string:
		return c, nil
	case bool:
		// lowercase, consistent with json.encode output
		if c {
			return "true", nil
		}
		return "false", nil
	case int:
		return strconv.Itoa(c), nil
	case float64:
		if math.IsNaN(c) || math.IsInf(c, 0) {
			return "", fmt.Errorf("float value %v is not representable in CSV", c)
		}
		// plain decimal notation, never scientific: 1000000.0 -> "1000000"
		return strconv.FormatFloat(c, 'f', -1, 64), nil
	case time.Time:
		return c.Format(time.RFC3339), nil
	default:
		// nested lists/dicts and anything else would serialize as Go syntax
		// ("[1 a]" / "map[a:1]") — reject loudly instead of corrupting
		return "", fmt.Errorf("unsupported cell type %T", v)
	}
}

// ModuleName defines the expected name for this Module when used in starlark's load() function, eg: load('csv', 'read_all')
const ModuleName = "csv"

var (
	once      sync.Once
	csvModule starlark.StringDict
)

// LoadModule loads the base64 module.
// It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		csvModule = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"read_all":       starlark.NewBuiltin(ModuleName+".read_all", readAll),
					"try_read_all":   starlark.NewBuiltin(ModuleName+".try_read_all", wrapTry(readAll)),
					"read_dict":      starlark.NewBuiltin(ModuleName+".read_dict", readDict),
					"try_read_dict":  starlark.NewBuiltin(ModuleName+".try_read_dict", wrapTry(readDict)),
					"write_all":      starlark.NewBuiltin(ModuleName+".write_all", writeAll),
					"try_write_all":  starlark.NewBuiltin(ModuleName+".try_write_all", wrapTry(writeAll)),
					"write_dict":     starlark.NewBuiltin(ModuleName+".write_dict", writeDict),
					"try_write_dict": starlark.NewBuiltin(ModuleName+".try_write_dict", wrapTry(writeDict)),
				},
			},
		}
	})
	return csvModule, nil
}

// wrapTry converts a builtin into its try_ variant: instead of aborting the
// whole script on failure, it returns a (value, error-string) pair with the
// Go error always nil — the shape established by lib/json's try_* functions.
// Argument-unpacking errors are captured the same way.
func wrapTry(fn dataconv.StarlarkFunc) dataconv.StarlarkFunc {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		res, err := fn(thread, b, args, kwargs)
		if err != nil {
			return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
		}
		if res == nil {
			res = starlark.None
		}
		return starlark.Tuple{res, starlark.None}, nil
	}
}

// readOptions holds the shared parameter set of the CSV reading builtins.
type readOptions struct {
	source                       tps.StringOrBytes
	comma, comment               starlark.String
	lazyQuotes, trimLeadingSpace bool
	fieldsPerRecord              int
	skipRow, limitRow            int
}

func unpackReadArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (*readOptions, error) {
	o := &readOptions{}
	if err := starlark.UnpackArgs(b.Name(), args, kwargs,
		"source", &o.source,
		"comma?", &o.comma,
		"comment", &o.comment,
		"lazy_quotes", &o.lazyQuotes,
		"trim_leading_space", &o.trimLeadingSpace,
		"fields_per_record?", &o.fieldsPerRecord,
		"skip?", &o.skipRow,
		"limit?", &o.limitRow); err != nil {
		return nil, err
	}
	return o, nil
}

// newReader builds the configured csv.Reader and consumes the skipped rows.
func (o *readOptions) newReader(b *starlark.Builtin) (*csv.Reader, error) {
	rawStr := file.TrimUTF8BOM([]byte(o.source.GoString()))
	csvr := csv.NewReader(replacecr.Reader(bytes.NewReader(rawStr)))
	csvr.LazyQuotes = o.lazyQuotes
	csvr.TrimLeadingSpace = o.trimLeadingSpace

	comma := string(o.comma)
	if comma == "" {
		comma = ","
	} else if len(comma) != 1 {
		return nil, fmt.Errorf("%s: expected comma param to be a single-character string", b.Name())
	}
	csvr.Comma = []rune(comma)[0]

	comment := string(o.comment)
	if comment != "" && len(comment) != 1 {
		return nil, fmt.Errorf("%s: expected comment param to be a single-character string", b.Name())
	} else if comment != "" {
		csvr.Comment = []rune(comment)[0]
	}
	csvr.FieldsPerRecord = o.fieldsPerRecord

	// pre-read to skip rows
	if o.skipRow > 0 {
		for i := 0; i < o.skipRow; i++ {
			if _, err := csvr.Read(); err != nil {
				return nil, fmt.Errorf("%s: %w", b.Name(), err)
			}
		}
		if o.fieldsPerRecord == 0 {
			// with the default fields_per_record=0 the first row read pins
			// the expected field count — rows consumed by skip must not be
			// that reference, the first kept row is
			csvr.FieldsPerRecord = 0
		}
	}
	return csvr, nil
}

// readAll gets all values from a csv source string.
func readAll(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	o, err := unpackReadArgs(b, args, kwargs)
	if err != nil {
		return nil, err
	}
	csvr, err := o.newReader(b)
	if err != nil {
		return starlark.None, err
	}

	// read rows one by one, so a positive limit stops both parsing and
	// memory growth at limit rows instead of materializing the whole file
	// first (which also aborted on malformed rows beyond the limit)
	var vals []starlark.Value
	for o.limitRow <= 0 || len(vals) < o.limitRow {
		rowStr, err := csvr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return starlark.None, fmt.Errorf("%s: %w", b.Name(), err)
		}
		row := make([]starlark.Value, len(rowStr))
		for j, cell := range rowStr {
			row[j] = starlark.String(cell)
		}
		vals = append(vals, starlark.NewList(row))
	}
	return starlark.NewList(vals), nil
}

// readDict reads a csv source string whose first row (after skip) is the
// header, returning a list of dicts keyed by the header fields.
func readDict(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	o, err := unpackReadArgs(b, args, kwargs)
	if err != nil {
		return nil, err
	}
	csvr, err := o.newReader(b)
	if err != nil {
		return starlark.None, err
	}

	// the first remaining row is the header; an empty source yields an
	// empty list rather than an error
	header, err := csvr.Read()
	if err == io.EOF {
		return starlark.NewList(nil), nil
	}
	if err != nil {
		return starlark.None, fmt.Errorf("%s: %w", b.Name(), err)
	}
	seen := make(map[string]bool, len(header))
	for _, h := range header {
		if seen[h] {
			return starlark.None, fmt.Errorf("%s: duplicate header field %q", b.Name(), h)
		}
		seen[h] = true
	}

	// limit counts data rows, the header is not included
	var vals []starlark.Value
	for o.limitRow <= 0 || len(vals) < o.limitRow {
		rowStr, err := csvr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return starlark.None, fmt.Errorf("%s: %w", b.Name(), err)
		}
		d := starlark.NewDict(len(header))
		for j, cell := range rowStr {
			if j >= len(header) {
				// only reachable with fields_per_record=-1; extra cells
				// have no field name to map to
				break
			}
			if err := d.SetKey(starlark.String(header[j]), starlark.String(cell)); err != nil {
				return starlark.None, fmt.Errorf("%s: %w", b.Name(), err)
			}
		}
		vals = append(vals, d)
	}
	return starlark.NewList(vals), nil
}

// writeAll writes a list of lists to a csv string.
func writeAll(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		buf   = &bytes.Buffer{}
		data  starlark.Value
		comma string
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "data", &data, "comma?", &comma); err != nil {
		return nil, err
	}

	// prepare writer
	csvw := csv.NewWriter(buf)
	if comma == "" {
		comma = ","
	} else if len(comma) != 1 {
		return starlark.None, fmt.Errorf("%s: expected comma param to be a single-character string", b.Name())
	}
	csvw.Comma = []rune(comma)[0]

	// convert data to [][]string
	val, err := dataconv.Unmarshal(data)
	if err != nil {
		return starlark.None, fmt.Errorf("%s: %w", b.Name(), err)
	}

	sl, ok := val.([]interface{})
	if !ok {
		return starlark.None, fmt.Errorf("%s: expected value to be an array type", b.Name())
	}

	var records [][]string
	for i, v := range sl {
		sl, ok := v.([]interface{})
		if !ok {
			return starlark.None, fmt.Errorf("%s: row %d is not an array type", b.Name(), i)
		}
		var row = make([]string, len(sl))
		for j, v := range sl {
			cell, err := cellToString(v)
			if err != nil {
				return starlark.None, fmt.Errorf("%s: row %d column %d: %w", b.Name(), i, j, err)
			}
			row[j] = cell
		}
		records = append(records, row)
	}

	// write all records
	if err := csvw.WriteAll(records); err != nil {
		return starlark.None, fmt.Errorf("%s: %w", b.Name(), err)
	}
	return starlark.String(buf.String()), nil
}

// writeDict writes a list of dictionaries to a csv string.
func writeDict(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		buf    = &bytes.Buffer{}
		data   starlark.Value
		header starlark.Iterable
		comma  string
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "data", &data, "header", &header, "comma?", &comma); err != nil {
		return nil, err
	}

	// prepare writer
	csvw := csv.NewWriter(buf)
	if comma == "" {
		comma = ","
	} else if len(comma) != 1 {
		return starlark.None, fmt.Errorf("%s: expected comma param to be a single-character string", b.Name())
	}
	csvw.Comma = []rune(comma)[0]

	// convert header to []string
	var headerStr []string
	iter := header.Iterate()
	defer iter.Done()
	var hv starlark.Value
	for iter.Next(&hv) {
		s, ok := starlark.AsString(hv)
		if !ok {
			return starlark.None, fmt.Errorf("%s: for parameter header: got %s, want string", b.Name(), hv.Type())
		}
		headerStr = append(headerStr, s)
	}
	if len(headerStr) == 0 {
		return starlark.None, fmt.Errorf("%s: header cannot be empty", b.Name())
	}

	// convert data to []map[string]interface{}
	val, err := dataconv.Unmarshal(data)
	if err != nil {
		return starlark.None, fmt.Errorf("%s: %w", b.Name(), err)
	}
	sl, ok := val.([]interface{})
	if !ok {
		return starlark.None, fmt.Errorf("%s: expected value to be an array type", b.Name())
	}

	// write header
	var records [][]string
	records = append(records, headerStr)
	for _, m := range sl {
		// cast to map
		mm, ok := m.(map[string]interface{})
		if !ok {
			return starlark.None, fmt.Errorf("%s: expected value to be a map type", b.Name())
		}
		// write row
		var row = make([]string, len(headerStr))
		for j, k := range headerStr {
			if v, ok := mm[k]; ok {
				cell, err := cellToString(v)
				if err != nil {
					return starlark.None, fmt.Errorf("%s: row %d field %q: %w", b.Name(), len(records)-1, k, err)
				}
				row[j] = cell
			}
		}
		records = append(records, row)
	}

	// write all records
	if err := csvw.WriteAll(records); err != nil {
		return starlark.None, fmt.Errorf("%s: %w", b.Name(), err)
	}
	return starlark.String(buf.String()), nil
}
