// Package csv reads comma-separated values from strings and writes CSV data to strings.
//
// Migrated from https://github.com/qri-io/starlib/tree/master/encoding/csv
package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"sync"

	"github.com/1set/starlet/dataconv"
	tps "github.com/1set/starlet/dataconv/types"
	"github.com/1set/starlet/internal/replacecr"
	"github.com/1set/starlet/lib/file"
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
					"read_all":   starlark.NewBuiltin(ModuleName+".read_all", readAll),
					"write_all":  starlark.NewBuiltin(ModuleName+".write_all", writeAll),
					"write_dict": starlark.NewBuiltin(ModuleName+".write_dict", writeDict),
				},
			},
		}
	})
	return csvModule, nil
}

// readAll gets all values from a csv source string.
func readAll(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		source                       tps.StringOrBytes
		lazyQuotes, trimLeadingSpace bool
		skipRow, limitRow            int
		fieldsPerRecord              int
		_comma, _comment             starlark.String
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs,
		"source", &source,
		"comma?", &_comma,
		"comment", &_comment,
		"lazy_quotes", &lazyQuotes,
		"trim_leading_space", &trimLeadingSpace,
		"fields_per_record?", &fieldsPerRecord,
		"skip?", &skipRow,
		"limit?", &limitRow); err != nil {
		return nil, err
	}

	// prepare reader
	rawStr := file.TrimUTF8BOM([]byte(source.GoString()))
	csvr := csv.NewReader(replacecr.Reader(bytes.NewReader(rawStr)))
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
	csvr.FieldsPerRecord = fieldsPerRecord

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
	vals := make([]starlark.Value, 0, len(strs))
	for i, rowStr := range strs {
		if limitRow > 0 && i >= limitRow {
			break
		}
		row := make([]starlark.Value, len(rowStr))
		for j, cell := range rowStr {
			row[j] = starlark.String(cell)
		}
		vals = append(vals, starlark.NewList(row))
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
				row[j] = fmt.Sprintf("%v", v)
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
