// Package csv reads comma-separated values from strings and writes CSV data to strings.
//
// Migrated from https://github.com/qri-io/starlib/tree/master/encoding/csv
package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"sync"

	"github.com/1set/starlet/dataconv"
	"github.com/1set/starlet/internal/replacecr"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

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
					"read_all":  starlark.NewBuiltin("read_all", readAll),
					"write_all": starlark.NewBuiltin("write_all", writeAll),
				},
			},
		}
	})
	return csvModule, nil
}

// readAll gets all values from a csv source string.
func readAll(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		source                       string
		lazyQuotes, trimLeadingSpace bool
		skipRow, limitRow            int
		fieldsPerRecord              int
		_comma, _comment             starlark.String
	)
	if err := starlark.UnpackArgs("read_all", args, kwargs,
		"source", &source,
		"comma?", &_comma,
		"comment", &_comment,
		"lazy_quotes", &lazyQuotes,
		"trim_leading_space", &trimLeadingSpace,
		"limit_column?", &fieldsPerRecord,
		"skip_row?", &skipRow,
		"limit_row?", &limitRow); err != nil {
		return nil, err
	}

	// prepare reader
	csvr := csv.NewReader(replacecr.Reader(strings.NewReader(source)))
	csvr.LazyQuotes = lazyQuotes
	csvr.TrimLeadingSpace = trimLeadingSpace

	comma := string(_comma)
	if comma == "" {
		comma = ","
	} else if len(comma) != 1 {
		return starlark.None, fmt.Errorf("%s: expected comma param to be a single-character string", b.Name())
	}
	csvr.Comma = []rune(comma)[0]

	comment := string(_comment)
	if comment != "" && len(comment) != 1 {
		return starlark.None, fmt.Errorf("%s: expected comment param to be a single-character string", b.Name())
	} else if comment != "" {
		csvr.Comment = []rune(comment)[0]
	}

	if fieldsPerRecord > 0 {
		csvr.FieldsPerRecord = fieldsPerRecord
	}

	// pre-read to skip rows
	if skipRow > 0 {
		for i := 0; i < skipRow; i++ {
			if _, err := csvr.Read(); err != nil {
				return starlark.None, fmt.Errorf("%s: %w", b.Name(), err)
			}
		}
	}

	// read all rows
	strs, err := csvr.ReadAll()
	if err != nil {
		return starlark.None, fmt.Errorf("%s: %w", b.Name(), err)
	}

	// convert and limit rows
	vals := make([]starlark.Value, len(strs))
	for i, rowStr := range strs {
		if limitRow > 0 && i >= limitRow {
			break
		}
		row := make([]starlark.Value, len(rowStr))
		for j, cell := range rowStr {
			row[j] = starlark.String(cell)
		}
		vals[i] = starlark.NewList(row)
	}
	return starlark.NewList(vals), nil
}

// writeAll writes a csv file to a string.
func writeAll(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		buf    = &bytes.Buffer{}
		source starlark.Value
		comma  string
	)

	if err := starlark.UnpackArgs("write_all", args, kwargs, "source", &source, "comma?", &comma); err != nil {
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

	// convert source to [][]string
	val, err := dataconv.Unmarshal(source)
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
			row[j] = fmt.Sprintf("%v", v)
		}
		records = append(records, row)
	}

	// write all records
	if err := csvw.WriteAll(records); err != nil {
		return starlark.None, fmt.Errorf("%s: %w", b.Name(), err)
	}
	return starlark.String(buf.String()), nil
}
