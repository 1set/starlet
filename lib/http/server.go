package http

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/1set/starlet/dataconv"
	itn "github.com/1set/starlet/internal"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	structNameRequest  = starlark.String("Request")
	structNameResponse = starlark.String("Response")
)

// ExportedServerRequest encapsulates HTTP request data in a format accessible to both Go code and Starlark scripts.
// This struct bridges Go's HTTP handling features with Starlark's dynamic scripting capabilities, enabling seamless
// interaction and manipulation of request properties in Go, while providing a structured, read-only view of the request
// data to Starlark scripts.
//
// Key Features:
//   - Full access to HTTP request properties (method, URL, headers, body) for reading and modification in Go.
//   - Structured, read-only representation of request data for Starlark scripts, enhancing scripting flexibility.
//   - JSON body support simplifies working with JSON payloads directly in scripts.
//
// Usage Pattern:
//   1. Convert an incoming http.Request to ExportedServerRequest with NewExportedServerRequest for access in Go.
//   2. Modify the ExportedServerRequest properties as needed in Go before handing off to Starlark.
//   3. Use the Struct method to convert the ExportedServerRequest to a Starlark struct, passing it to Starlark scripts
//      for read-only access. This step allows scripts to inspect the request's properties.
//   4. Since the Starlark struct is read-only, modifications to the request must be performed in Go, either before
//      or after script execution.
//
// This design prioritizes ease of use, security, and performance, facilitating dynamic and complex request processing
// logic through Go and Starlark. It ensures the integrity of the HTTP request handling by preventing unauthorized
// modifications and protecting against potential security threats. Developers are encouraged to validate all modifications
// and interactions with the request data to maintain the server's security posture.
type ExportedServerRequest struct {
	Method   string         // The HTTP method (e.g., GET, POST, PUT, DELETE)
	URL      *url.URL       // The request URL
	Proto    string         // The protocol used for the request (e.g., HTTP/1.1)
	Host     string         // The host specified in the request
	Remote   string         // The remote address of the client
	Header   http.Header    // The HTTP headers included in the request
	Encoding []string       // The transfer encodings specified in the request
	Body     []byte         // The request body data
	JSONData starlark.Value // The request body data as Starlark value
}

// NewExportedServerRequest creates a new ExportedServerRequest from an http.Request.
func NewExportedServerRequest(r *http.Request) (*ExportedServerRequest, error) {
	if r == nil {
		return nil, errors.New("nil request")
	}

	// attempt to read the request body
	var (
		err  error
		body []byte
		sv   starlark.Value = starlark.None
	)
	if r.Body != nil {
		// request like faked GET may not have body
		if body, err = ioutil.ReadAll(r.Body); err != nil {
			return nil, err
		}

		// reset the request body to allow multiple reads
		_ = r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}

	// marshal body as json
	if len(body) > 0 {
		if sv, err = dataconv.UnmarshalStarlarkJSON(body); err != nil {
			sv = starlark.None
		}
	}

	// create the exported request
	return &ExportedServerRequest{
		Method:   r.Method,
		URL:      r.URL,
		Proto:    r.Proto,
		Host:     r.Host,
		Remote:   r.RemoteAddr,
		Header:   r.Header,
		Encoding: r.TransferEncoding,
		Body:     body,
		JSONData: sv,
	}, nil
}

// Struct returns a Starlark struct representation of the ExportedServerRequest, which exposes the following fields to Starlark scripts:
//   - method: The HTTP method (e.g., GET, POST, PUT, DELETE)
//   - url: The request URL
//   - proto: The protocol used for the request (e.g., HTTP/1.1)
//   - host: The host specified in the request
//   - remote: The remote address of the client
//   - header: The HTTP headers included in the request
//   - query: The query parameters included in the request URL
//   - encoding: The transfer encodings specified in the request
//   - body: The request body data
//   - json: The request body data as Starlark value, if it is valid JSON or None otherwise
func (r *ExportedServerRequest) Struct() *starlarkstruct.Struct {
	// prepare struct members
	sd := starlark.StringDict{
		"method":   starlark.String(r.Method),
		"url":      starlark.String(r.URL.String()),
		"proto":    starlark.String(r.Proto),
		"host":     starlark.String(r.Host),
		"remote":   starlark.String(r.Remote),
		"header":   mapStrs2Dict(r.Header),
		"query":    mapStrs2Dict(r.URL.Query()),
		"encoding": sliceStr2List(r.Encoding),
		"body":     starlark.String(r.Body),
		"json":     r.JSONData,
	}
	// create struct
	return starlarkstruct.FromStringDict(structNameRequest, sd)
}

// Write writes the request data back to a provided http.Request instance.
func (r *ExportedServerRequest) Write(req *http.Request) (err error) {
	if req == nil {
		return errors.New("nil request")
	}

	// set request method, URL, and protocol
	req.Method = r.Method
	req.URL = r.URL
	req.Proto = r.Proto
	req.Host = r.Host
	req.RemoteAddr = r.Remote
	req.Header = r.Header
	req.TransferEncoding = r.Encoding
	// set request body
	req.Body = ioutil.NopCloser(bytes.NewReader(r.Body))

	return err
}

// ConvertServerRequest converts a http.Request to a Starlark struct for use in Starlark scripts on the server side.
func ConvertServerRequest(r *http.Request) *starlarkstruct.Struct {
	sr, err := NewExportedServerRequest(r)
	if err != nil || sr == nil {
		return nil
	}
	return sr.Struct()
}

// NewServerResponse creates a new ServerResponse.
func NewServerResponse() *ServerResponse {
	return &ServerResponse{}
}

// ServerResponse is a struct that enables HTTP response manipulation within Starlark scripts,
// facilitating dynamic preparation of HTTP responses in Go-based web servers executing such scripts.
//
// Key Features:
//   - Setting HTTP status codes.
//   - Adding and managing HTTP headers.
//   - Specifying the content type of the response.
//   - Setting the response body with support for various data types (e.g., binary, text, HTML, JSON).
//
// Usage Pattern:
//   1. Create a ServerResponse instance using NewServerResponse().
//   2. Utilize the Struct() method to obtain a Starlark struct that exposes ServerResponse functionalities to Starlark scripts.
//   3. In the Starlark script, utilize provided methods (e.g., set_status, add_header, set_content_type) to prepare the response.
//   4. Back in Go, the ServerResponse instance can directly write its content to an http.ResponseWriter using its Write() method.
//      Alternatively, you can call the Export() method to convert the ServerResponse into an ExportedServerResponse for modification,
//      which is then capable of being written to an http.ResponseWriter using its Write() method.
//
// Internally, ServerResponse uses a private contentDataType enum to manage the intended type of the response data,
// allowing for automatic adjustment of the Content-Type header based on the set data type by the Starlark script.
//
// The ExportedServerResponse struct simplifies ServerResponse for interoperability with Go's standard http package,
// comprising an HTTP status code, headers, and data for the HTTP response. Its Write() method allows for the prepared
// response to be efficiently written to an http.ResponseWriter, ensuring correct header setting and response body data writing.
//
// Note: Direct manipulation of ServerResponse and its methods by Starlark scripts necessitates validation of script inputs
// to mitigate potential security issues like header injection attacks. This design allows scripts to dynamically prepare
// HTTP responses while maintaining a secure and controlled server environment.
type ServerResponse struct {
	statusCode  int
	headers     map[string][]string
	contentType string
	dataType    contentDataType
	data        []byte
}

// Struct returns a Starlark struct representation of the ServerResponse, which exposes the following methods to Starlark scripts:
//   - set_status(code): Sets the HTTP status code for the response.
//   - set_code(code): An alias for set_status.
//   - add_header(key, value): Adds a header with the given key and value to the response.
//   - set_content_type(contentType): Sets the Content-Type header for the response.
//   - set_data(data): Sets the response data as binary data.
//   - set_json(data): Sets the response data as JSON, marshaling the given Starlark value to JSON.
//   - set_text(data): Sets the response data as plain text.
//   - set_html(data): Sets the response data as HTML.
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
