package http

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/1set/starlet/dataconv"
	itn "github.com/1set/starlet/internal"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	structNameRequest  = starlark.String("Request")
	structNameResponse = starlark.String("Response")
)

// ConvertServerRequest converts a http.Request to a Starlark struct for use in Starlark scripts on the server side.
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
	return starlarkstruct.FromStringDict(structNameRequest, sd)
}

// NewServerResponse creates a new ServerResponse.
func NewServerResponse() *ServerResponse {
	return &ServerResponse{}
}

// ServerResponse is a Starlark struct to save info in Starlark scripts to modify http.ResponseWriter outside on the server side.
type ServerResponse struct {
	statusCode  int
	headers     map[string][]string
	contentType string
	dataType    contentDataType
	data        []byte
}

// Struct returns a Starlark struct for use in Starlark scripts.
func (r *ServerResponse) Struct() *starlarkstruct.Struct {
	// prepare struct members
	sd := starlark.StringDict{
		"set_status":       starlark.NewBuiltin("set_status", r.setStatus),
		"set_code":         starlark.NewBuiltin("set_code", r.setStatus), // alias for set_status
		"add_header":       starlark.NewBuiltin("add_header", r.addHeaderValue),
		"set_content_type": starlark.NewBuiltin("set_content_type", r.setContentType),
		"set_data":         starlark.NewBuiltin("set_data", r.setData(contentDataBinary)),
		"set_json":         starlark.NewBuiltin("set_json", r.setJSONData),
		"set_text":         starlark.NewBuiltin("set_text", r.setData(contentDataText)),
		"set_html":         starlark.NewBuiltin("set_html", r.setData(contentDataHTML)),
	}
	// create struct
	return starlarkstruct.FromStringDict(structNameResponse, sd)
}

// ExportedServerResponse is a struct to export the response data to Go.
type ExportedServerResponse struct {
	StatusCode int         // StatusCode is the status code of the response.
	Header     http.Header // Header is the header of the response, a map of string to list of strings. Content-Type is set automatically.
	Data       []byte      // Data is the data of the response, usually the body content.
}

func (d *ExportedServerResponse) Write(w http.ResponseWriter) (err error) {
	// basic check
	if w == nil {
		err = errors.New("nil response writer")
		return
	}
	if d == nil {
		err = errors.New("nil exported response")
		return
	}

	// write header first, and then status code & data
	copyHeader(w.Header(), d.Header)
	w.WriteHeader(d.StatusCode)
	if d.Data != nil {
		_, err = w.Write(d.Data)
	}
	return
}

// Write writes the response to http.ResponseWriter.
func (r *ServerResponse) Write(w http.ResponseWriter) (err error) {
	d := r.Export()
	return d.Write(w)
}

// Export dumps the response data to a struct for later use in Go.
func (r *ServerResponse) Export() *ExportedServerResponse {
	resp := ExportedServerResponse{
		Header: make(http.Header, len(r.headers)),
		Data:   r.data,
	}

	// status code
	if r.statusCode > 0 {
		resp.StatusCode = r.statusCode
	} else {
		resp.StatusCode = http.StatusOK
	}

	// headers
	if r.headers != nil {
		for k, vs := range r.headers {
			for _, v := range vs {
				resp.Header.Add(k, v)
			}
		}
	}

	// content type
	contentType := r.contentType
	if contentType == "" {
		switch r.dataType {
		case contentDataJSON:
			contentType = "application/json"
		case contentDataText:
			contentType = "text/plain"
		case contentDataHTML:
			contentType = "text/html"
		case contentDataBinary:
			fallthrough
		default:
			contentType = "application/octet-stream"
		}
	}
	resp.Header.Set("Content-Type", contentType)

	// for later use
	return &resp
}

func (r *ServerResponse) setStatus(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var code uint16
	if err := starlark.UnpackPositionalArgs(b.Name(), args, nil, 1, &code); err != nil {
		return nil, err
	}
	if code < 100 || code > 599 {
		return nil, errors.New("invalid status code")
	}
	r.statusCode = int(code)
	return starlark.None, nil
}

func (r *ServerResponse) addHeaderValue(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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

func (r *ServerResponse) setContentType(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var ct itn.StringOrBytes
	if err := starlark.UnpackPositionalArgs(b.Name(), args, nil, 1, &ct); err != nil {
		return nil, err
	}
	r.contentType = ct.GoString()
	return starlark.None, nil
}

// setData sets the response data with the given type except JSON.
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

// setJSONData marshals the given Starlark value to JSON and sets the response data.
func (r *ServerResponse) setJSONData(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var data starlark.Value
	if err := starlark.UnpackPositionalArgs(b.Name(), args, nil, 1, &data); err != nil {
		return nil, err
	}
	// convert to JSON
	bs, err := dataconv.MarshalStarlarkJSON(data, 0)
	if err != nil {
		return nil, err
	}
	// set data
	r.data = []byte(bs)
	r.dataType = contentDataJSON
	return starlark.None, nil
}

type contentDataType uint

const (
	contentDataBinary contentDataType = iota
	contentDataJSON
	contentDataText
	contentDataHTML
)

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

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			//dst.Add(k, v)
			dst[k] = append(dst[k], v)
		}
	}
}
