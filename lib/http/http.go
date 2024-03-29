// Package http defines a module for doing http operations in Starlark.
//
// Migrated from: https://github.com/qri-io/starlib/tree/master/http
package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/1set/starlet/dataconv"
	itn "github.com/1set/starlet/internal"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used
// in starlark's load() function, eg: load('http', 'get')
const ModuleName = "http"

var (
	// UserAgent is the default user agent for http requests, override with a custom value before calling LoadModule.
	UserAgent = "Starlet-http-client/" + itn.StarletVersion
	// TimeoutSecond is the default timeout in seconds for http requests, override with a custom value before calling LoadModule.
	TimeoutSecond = 30
	// SkipInsecureVerify controls whether to skip TLS verification, override with a custom value before calling LoadModule.
	SkipInsecureVerify = false
	// DisableRedirect controls whether to follow redirects, override with a custom value before calling LoadModule.
	DisableRedirect = false
	// Client is the http client used to create the http module, override with a custom client before calling LoadModule.
	Client *http.Client
	// Guard is a global RequestGuard used in LoadModule, override with a custom implementation before calling LoadModule.
	Guard RequestGuard
)

// RequestGuard controls access to http by checking before making requests
// if Allowed returns an error the request will be denied
type RequestGuard interface {
	Allowed(thread *starlark.Thread, req *http.Request) (*http.Request, error)
}

// LoadModule creates an http Module
func LoadModule() (starlark.StringDict, error) {
	var m = &Module{}
	if Client != nil {
		m.cli = Client
	}
	if Guard != nil {
		m.rg = Guard
	}
	ns := starlark.StringDict{
		"http": m.Struct(),
	}
	return ns, nil
}

// Module joins http tools to a dataset, allowing dataset
// to follow along with http requests
type Module struct {
	cli *http.Client
	rg  RequestGuard
}

// Struct returns this module's methods as a starlark Struct
func (m *Module) Struct() *starlarkstruct.Struct {
	return starlarkstruct.FromStringDict(starlarkstruct.Default, m.StringDict())
}

// StringDict returns all module methods in a starlark.StringDict
func (m *Module) StringDict() starlark.StringDict {
	methods := []string{"get", "put", "post", "postForm", "delete", "head", "patch", "options"}
	sd := make(starlark.StringDict, len(methods))
	for _, name := range methods {
		sd[name] = starlark.NewBuiltin(ModuleName+"."+name, m.reqMethod(name))
	}
	return sd
}

// reqMethod is a factory function for generating starlark builtin functions for different http request methods
func (m *Module) reqMethod(method string) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			urlv          starlark.String
			params        = &starlark.Dict{}
			headers       = &starlark.Dict{}
			formBody      = &starlark.Dict{}
			formEncoding  starlark.String
			auth          starlark.Tuple
			body          itn.StringOrBytes
			jsonBody      starlark.Value
			timeout       = itn.FloatOrInt(TimeoutSecond)
			allowRedirect = starlark.Bool(!DisableRedirect)
			verifySSL     = starlark.Bool(!SkipInsecureVerify)
		)

		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "url", &urlv, "params?", &params, "headers", &headers, "body", &body, "form_body", &formBody, "form_encoding", &formEncoding, "json_body", &jsonBody,
			"auth", &auth, "timeout?", &timeout, "allow_redirects?", &allowRedirect, "verify?", &verifySSL); err != nil {
			return nil, err
		}

		rawURL, err := AsString(urlv)
		if err != nil {
			return nil, err
		}
		if err = setQueryParams(&rawURL, params); err != nil {
			return nil, err
		}

		req, err := http.NewRequest(strings.ToUpper(method), rawURL, nil)
		if err != nil {
			return nil, err
		}
		if m.rg != nil {
			req, err = m.rg.Allowed(thread, req)
			if err != nil {
				return nil, err
			}
		}

		if err = setHeaders(req, headers); err != nil {
			return nil, err
		}
		if err = setAuth(req, auth); err != nil {
			return nil, err
		}
		if err = setBody(req, body.StarlarkString(), formBody, formEncoding, jsonBody); err != nil {
			return nil, err
		}

		cli := m.getHTTPClient(float64(timeout), bool(allowRedirect), bool(verifySSL))
		res, err := cli.Do(req)
		if err != nil {
			return nil, err
		}

		r := &Response{*res}
		return r.Struct(), nil
	}
}

func (m *Module) getHTTPClient(timeoutSec float64, allowRedirect, verifySSL bool) *http.Client {
	// return existing client if set
	if m.cli != nil {
		return m.cli
	}
	// set timeout to 30 seconds if it's negative
	if timeoutSec < 0 {
		timeoutSec = 30
	}
	cli := &http.Client{Timeout: time.Duration(timeoutSec * float64(time.Second))}
	// skip TLS verification if set
	if !verifySSL {
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.TLSClientConfig.InsecureSkipVerify = true
		cli.Transport = tr
	}
	// disable redirects if set
	if !allowRedirect {
		cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	return cli
}

// AsString unquotes a starlark string value
func AsString(x starlark.Value) (string, error) {
	return strconv.Unquote(x.String())
}

// Encodings for form data.
// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/POST
const (
	formEncodingMultipart = "multipart/form-data"
	formEncodingURL       = "application/x-www-form-urlencoded"
)

func setQueryParams(rawurl *string, params *starlark.Dict) error {
	keys := params.Keys()
	if len(keys) == 0 {
		return nil
	}

	u, err := url.Parse(*rawurl)
	if err != nil {
		return err
	}

	q := u.Query()
	for _, key := range keys {
		keystr, err := AsString(key)
		if err != nil {
			return err
		}

		val, _, err := params.Get(key)
		if err != nil {
			return err
		}
		if val.Type() != "string" {
			return fmt.Errorf("expected param value for key '%s' to be a string. got: '%s'", key, val.Type())
		}
		valstr, err := AsString(val)
		if err != nil {
			return err
		}

		q.Set(keystr, valstr)
	}

	u.RawQuery = q.Encode()
	*rawurl = u.String()
	return nil
}

func setAuth(req *http.Request, auth starlark.Tuple) error {
	if len(auth) == 0 {
		return nil
	} else if len(auth) == 2 {
		username, err := AsString(auth[0])
		if err != nil {
			return fmt.Errorf("parsing auth username string: %s", err.Error())
		}
		password, err := AsString(auth[1])
		if err != nil {
			return fmt.Errorf("parsing auth password string: %s", err.Error())
		}
		req.SetBasicAuth(username, password)
		return nil
	}
	return fmt.Errorf("expected two values for auth params tuple")
}

func setHeaders(req *http.Request, headers *starlark.Dict) error {
	var (
		keys    = headers.Keys()
		UAKey   = "User-Agent"
		isUASet = false
	)
	for _, key := range keys {
		keystr, err := AsString(key)
		if err != nil {
			return err
		}

		val, _, err := headers.Get(key)
		if err != nil {
			return err
		}
		if val.Type() != "string" {
			return fmt.Errorf("expected param value for key '%s' to be a string. got: '%s'", key, val.Type())
		}
		valstr, err := AsString(val)
		if err != nil {
			return err
		}

		req.Header.Add(keystr, valstr)
		if keystr == UAKey {
			isUASet = true
		}
	}

	if UserAgent != "" && !isUASet {
		req.Header.Set(UAKey, UserAgent)
	}
	return nil
}

func setBody(req *http.Request, body starlark.String, formData *starlark.Dict, formEncoding starlark.String, jsondata starlark.Value) error {
	if !dataconv.IsEmptyString(body) {
		uq, err := strconv.Unquote(body.String())
		if err != nil {
			return err
		}
		req.Body = ioutil.NopCloser(strings.NewReader(uq))
		// Specifying the Content-Length ensures that https://go.dev/src/net/http/transfer.go doesnt specify Transfer-Encoding: chunked which is not supported by some endpoints.
		// This is required when using ioutil.NopCloser method for the request body (see ShouldSendChunkedRequestBody() in the library mentioned above).
		req.ContentLength = int64(len(uq))

		return nil
	}

	if jsondata != nil && jsondata.String() != "" {
		req.Header.Set("Content-Type", "application/json")
		data, err := dataconv.MarshalStarlarkJSON(jsondata, 0)
		if err != nil {
			return err
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(data)))
		req.ContentLength = int64(len(data))
	}

	if formData != nil && formData.Len() > 0 {
		form := url.Values{}
		for _, key := range formData.Keys() {
			keystr, err := AsString(key)
			if err != nil {
				return err
			}

			val, _, err := formData.Get(key)
			if err != nil {
				return err
			}
			if val.Type() != "string" {
				return fmt.Errorf("expected param value for key '%s' to be a string. got: '%s'", key, val.Type())
			}
			valstr, err := AsString(val)
			if err != nil {
				return err
			}

			form.Add(keystr, valstr)
		}

		var contentType string
		switch formEncoding {
		case formEncodingURL, "":
			contentType = formEncodingURL
			req.Body = ioutil.NopCloser(strings.NewReader(form.Encode()))
			req.ContentLength = int64(len(form.Encode()))

		case formEncodingMultipart:
			var b bytes.Buffer
			mw := multipart.NewWriter(&b)
			defer mw.Close()

			contentType = mw.FormDataContentType()

			for k, values := range form {
				for _, v := range values {
					w, err := mw.CreateFormField(k)
					if err != nil {
						return err
					}
					if _, err := w.Write([]byte(v)); err != nil {
						return err
					}
				}
			}

			req.Body = ioutil.NopCloser(&b)

		default:
			return fmt.Errorf("unknown form encoding: %s", formEncoding)
		}

		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", contentType)
		}
	}

	return nil
}

// Response represents an HTTP response, wrapping a go http.Response with
// starlark methods
type Response struct {
	http.Response
}

// Struct turns a response into a *starlark.Struct
func (r *Response) Struct() *starlarkstruct.Struct {
	return starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"url":         starlark.String(r.Request.URL.String()),
		"status_code": starlark.MakeInt(r.StatusCode),
		"headers":     r.HeadersDict(),
		"encoding":    starlark.String(strings.Join(r.TransferEncoding, ",")),
		"body":        starlark.NewBuiltin("body", r.Text),
		"json":        starlark.NewBuiltin("json", r.JSON),
	})
}

// HeadersDict flops
func (r *Response) HeadersDict() *starlark.Dict {
	d := new(starlark.Dict)
	for key, vals := range r.Header {
		if err := d.SetKey(starlark.String(key), starlark.String(strings.Join(vals, ","))); err != nil {
			panic(err)
		}
	}
	return d
}

// Text returns the raw data as a string
func (r *Response) Text(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// reset reader to allow multiple calls
	_ = r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewReader(data))

	// wraps as result
	return starlark.String(data), nil
}

// JSON attempts to parse the response body as JSON
func (r *Response) JSON(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// reset reader to allow multiple calls
	_ = r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewReader(body))

	// use internal marshaler to support starlark types
	return dataconv.UnmarshalStarlarkJSON(body)
}
