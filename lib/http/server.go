package http

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/1set/starlet/dataconv"
	itn "github.com/1set/starlet/internal"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ConvertServerRequest converts a http.Request to a Starlark struct for use in Starlark scripts.
func ConvertServerRequest(r *http.Request) *starlarkstruct.Struct {
	// for nil request, return nil
	if r == nil {
		return nil
	}

	// prepare struct members
	sd := starlark.StringDict{
		"body": starlark.None,
		"json": starlark.None,
	}

	// set headers, query, and other fields
	sd["host"] = starlark.String(r.Host)
	sd["remote"] = starlark.String(r.RemoteAddr)
	sd["url"] = starlark.String(r.URL.String())
	sd["proto"] = starlark.String(r.Proto)
	sd["method"] = starlark.String(r.Method)
	sd["header"] = mapStrs2Dict(r.Header)
	sd["query"] = mapStrs2Dict(r.URL.Query())
	sd["encoding"] = sliceStr2List(r.TransferEncoding)

	// for body content
	var (
		bs  []byte
		err error
	)
	if r.Body != nil {
		if bs, err = ioutil.ReadAll(r.Body); err == nil {
			sd["body"] = starlark.String(bs)
			// reset reader to allow multiple calls outside
			_ = r.Body.Close()
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bs))
		}
	}
	if bs != nil {
		if sv, err := dataconv.UnmarshalStarlarkJSON(bs); err == nil {
			sd["json"] = sv
		}
	}

	// create struct
	return starlarkstruct.FromStringDict(starlark.String("Request"), sd)
}

func mapStrs2Dict(m map[string][]string) *starlark.Dict {
	d := &starlark.Dict{}
	for k, v := range m {
		_ = d.SetKey(starlark.String(k), sliceStr2List(v))
	}
	return d
}

func sliceStr2List(s []string) *starlark.List {
	l := make([]starlark.Value, len(s))
	for i, v := range s {
		l[i] = starlark.String(v)
	}
	return starlark.NewList(l)
}

type contentDataType uint

const (
	contentDataBinary contentDataType = iota
	contentDataJSON
	contentDataText
	contentDataHTML
)

// ServerResponse is a Starlark struct to save info in Starlark scripts to modify http.ResponseWriter outside.
type ServerResponse struct {
	statusCode  int
	headers     map[string][]string
	contentType string
	dataType    contentDataType
	data        []byte
}

func (r *ServerResponse) setStatus(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var code uint8
	if err := starlark.UnpackPositionalArgs(b.Name(), args, nil, 1, &code); err != nil {
		return nil, err
	}
	r.statusCode = int(code)
	return starlark.None, nil
}

func (r *ServerResponse) setData(dt contentDataType) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var data itn.StringOrBytes
		if err := starlark.UnpackPositionalArgs(b.Name(), args, nil, 1, &data); err != nil {
			return nil, err
		}
		r.dataType = dt
		r.data = data.GoBytes()
		return starlark.None, nil
	}
}

func (r *ServerResponse) setContentType(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var ct itn.StringOrBytes
	if err := starlark.UnpackPositionalArgs(b.Name(), args, nil, 1, &ct); err != nil {
		return nil, err
	}
	r.contentType = ct.GoString()
	return starlark.None, nil
}

func (r *ServerResponse) setHeader(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key, value itn.StringOrBytes
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "key", &key, "value", &value); err != nil {
		return nil, err
	}
	k, v := key.GoString(), value.GoString()
	if r.headers == nil {
		r.headers = make(map[string][]string)
	}
	r.headers[k] = append(r.headers[k], v)
	return starlark.None, nil
}
