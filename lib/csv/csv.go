// Package csv reads comma-separated values from strings and writes CSV data to strings.
//
// Migrated from https://github.com/qri-io/starlib/tree/master/encoding/csv
package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/1set/starlet/dataconv"
	"github.com/1set/starlet/internal/replacecr"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"io"
	"strings"
	"sync"
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
func readAll(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		r io.Reader

		source                       starlark.Value
		lazyQuotes, trimLeadingSpace starlark.Bool
		skip                         = starlark.MakeInt(0)
		fieldsPerRecord              = starlark.MakeInt(0)
		_comma, _comment             starlark.String
	)
	err := starlark.UnpackArgs("read_all", args, kwargs,
		"source", &source,
		"comma?", &_comma,
		"comment", &_comment,
		"lazy_quotes", &lazyQuotes,
		"trim_leading_space", &trimLeadingSpace,
		"fields_per_record", &fieldsPerRecord,
		"skip", &skip)

	if err != nil {
		return nil, err
	}

	switch source.Type() {
	case "string":
		str := string(source.(starlark.String))
		r = strings.NewReader(str)
	}
	csvr := csv.NewReader(replacecr.Reader(r))
	csvr.LazyQuotes = bool(lazyQuotes)
	csvr.TrimLeadingSpace = bool(trimLeadingSpace)

	comma := string(_comma)
	if comma == "" {
		comma = ","
	} else if len(comma) != 1 {
		return starlark.None, fmt.Errorf("expected comma param to be a single-character string")
	}
	csvr.Comma = []rune(comma)[0]

	comment := string(_comment)
	if comment != "" && len(comment) != 1 {
		return starlark.None, fmt.Errorf("expected comment param to be a single-character string")
	} else if comment != "" {
		csvr.Comment = []rune(comment)[0]
	}

	if fpr, ok := fieldsPerRecord.Int64(); ok && fpr != 0 {
		csvr.FieldsPerRecord = int(fpr)
	}

	if s, ok := skip.Int64(); ok {
		for i := 0; i < int(s); i++ {
			if _, err := csvr.Read(); err != nil {
				return starlark.None, err
			}
		}
	}

	strs, err := csvr.ReadAll()
	if err != nil {
		return starlark.None, err
	}

	vals := make([]starlark.Value, len(strs))
	for i, rowStr := range strs {
		row := make([]starlark.Value, len(rowStr))
		for j, cell := range rowStr {
			row[j] = starlark.String(cell)
		}
		vals[i] = starlark.NewList(row)
	}
	return starlark.NewList(vals), nil
}

// writeAll writes a csv file to a string.
func writeAll(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		buf = &bytes.Buffer{}

		source starlark.Value
		_comma starlark.String
	)

	if err := starlark.UnpackArgs("write_all", args, kwargs, "source", &source, "comma?", &_comma); err != nil {
		return nil, err
	}

	csvw := csv.NewWriter(buf)
	comma := string(_comma)
	if comma == "" {
		comma = ","
	} else if len(comma) != 1 {
		return starlark.None, fmt.Errorf("expected comma param to be a single-character string")
	}
	csvw.Comma = []rune(comma)[0]

	val, err := dataconv.Unmarshal(source)
	if err != nil {
		return starlark.None, err
	}

	sl, ok := val.([]interface{})
	if !ok {
		return starlark.None, fmt.Errorf("expected value to be an array type")
	}

	var records [][]string
	for i, v := range sl {
		sl, ok := v.([]interface{})
		if !ok {
			return starlark.None, fmt.Errorf("row %d is not an array type", i)
		}
		var row = make([]string, len(sl))
		for j, v := range sl {
			row[j] = fmt.Sprintf("%v", v)
		}
		records = append(records, row)
	}

	if err := csvw.WriteAll(records); err != nil {
		return starlark.None, err
	}
	return starlark.String(buf.String()), nil
}
