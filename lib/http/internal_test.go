package http

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/1set/starlet/dataconv/types"
	"go.starlark.net/starlark"
)

// postFormTransport keeps request dispatch tests independent of networking.
type postFormTransport func(*http.Request) (*http.Response, error)

func (f postFormTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

type postFormGuard func(*starlark.Thread, *http.Request) (*http.Request, error)

func (f postFormGuard) Allowed(thread *starlark.Thread, req *http.Request) (*http.Request, error) {
	return f(thread, req)
}

func TestPostFormDispatch(t *testing.T) {
	for _, tc := range []struct {
		name, builtin string
		method        starlark.Value
		urlEncoded    bool
	}{
		{"direct", "postForm", nil, true},
		{"try_direct", "try_postForm", nil, true},
		{"call_camel", "call", starlark.String("postForm"), true},
		{"call_lower", "call", starlark.String("postform"), true},
		{"call_upper", "call", starlark.String("POSTFORM"), true},
		{"call_bytes", "call", starlark.Bytes("PoStFoRm"), true},
		{"try_call", "try_call", starlark.String("POSTFORM"), true},
		{"ordinary_post", "call", starlark.String("POST"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := 0
			m := NewModule()
			m.SetClient(&http.Client{Transport: postFormTransport(func(req *http.Request) (*http.Response, error) {
				requests++
				if req.Method != "POST" {
					t.Errorf("method = %q, want POST", req.Method)
				}
				if tc.urlEncoded {
					if got := req.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
						t.Errorf("request %d content type = %q", requests, got)
					}
					body, err := ioutil.ReadAll(req.Body)
					if err != nil || string(body) != "a=b+c" {
						t.Errorf("body = %q, error = %v", body, err)
					}
				} else if !strings.HasPrefix(req.Header.Get("Content-Type"), "multipart/form-data;") {
					t.Errorf("ordinary POST lost explicit multipart encoding")
				}
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: ioutil.NopCloser(strings.NewReader("")), Request: req}, nil
			})})
			fn := m.StringDict()[tc.builtin]
			form := starlark.NewDict(1)
			if err := form.SetKey(starlark.String("a"), starlark.String("b c")); err != nil {
				t.Fatal(err)
			}
			args := starlark.Tuple{starlark.String("https://example.test/form")}
			if tc.method != nil {
				args = append(starlark.Tuple{tc.method}, args...)
			}
			kwargs := []starlark.Tuple{{starlark.String("form_body"), form}, {starlark.String("form_encoding"), starlark.String("multipart/form-data")}}
			// Reuse the same exported callable: postForm must force its
			// encoding on every invocation, including after a prior request.
			for i := 0; i < 2; i++ {
				got, err := starlark.Call(&starlark.Thread{}, fn, args, kwargs)
				if err != nil {
					t.Fatalf("call %d: %v", i+1, err)
				}
				if strings.HasPrefix(tc.builtin, "try_") {
					pair, ok := got.(starlark.Tuple)
					if !ok || len(pair) != 2 || pair[0] == starlark.None || pair[1] != starlark.None {
						t.Fatalf("try result = %v, want (response, None)", got)
					}
				} else if got == nil || got == starlark.None {
					t.Fatalf("result = %v, want response", got)
				}
				if requests != i+1 {
					t.Fatalf("requests = %d, want %d", requests, i+1)
				}
			}
		})
	}
}

func TestPostFormRequestGuard(t *testing.T) {
	for _, name := range []string{"postForm", "try_postForm", "call", "try_call"} {
		t.Run(name, func(t *testing.T) {
			m := NewModule()
			denied := errors.New("form requests denied")
			guardCalls := 0
			m.SetGuard(postFormGuard(func(_ *starlark.Thread, req *http.Request) (*http.Request, error) {
				guardCalls++
				if req.Method != "POST" {
					t.Errorf("guard sees method %q, want POST", req.Method)
				}
				return nil, denied
			}))
			m.SetClient(&http.Client{Transport: postFormTransport(func(*http.Request) (*http.Response, error) {
				t.Error("denied request reached transport")
				return nil, errors.New("unexpected transport call")
			})})
			args := starlark.Tuple{starlark.String("https://example.test/form")}
			if name == "call" || name == "try_call" {
				args = append(starlark.Tuple{starlark.String("PoStFoRm")}, args...)
			}
			got, err := starlark.Call(&starlark.Thread{}, m.StringDict()[name], args, nil)
			if strings.HasPrefix(name, "try_") {
				pair, ok := got.(starlark.Tuple)
				if err != nil || !ok || len(pair) != 2 || pair[0] != starlark.None || pair[1] != starlark.String(denied.Error()) {
					t.Errorf("try result = %v, error = %v", got, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), denied.Error()) {
				t.Errorf("error = %v, want guard denial", err)
			}
			if guardCalls != 1 {
				t.Errorf("guard called %d times, want 1", guardCalls)
			}
		})
	}
}

func BenchmarkRequestDispatch(b *testing.B) {
	for _, tc := range []struct{ name, builtin, method string }{
		{"direct_get", "get", ""},
		{"call_get", "call", "GET"},
		{"call_post", "call", "POST"},
		{"direct_postform", "postForm", ""},
	} {
		b.Run(tc.name, func(b *testing.B) {
			m := NewModule()
			m.SetClient(&http.Client{Transport: postFormTransport(func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: ioutil.NopCloser(strings.NewReader("")), Request: req}, nil
			})})
			fn := m.StringDict()[tc.builtin]
			args := starlark.Tuple{starlark.String("https://example.test/form")}
			if tc.method != "" {
				args = append(starlark.Tuple{starlark.String(tc.method)}, args...)
			}
			thread := &starlark.Thread{}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := starlark.Call(thread, fn, args, nil); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// we're ok with testing private functions if it simplifies the test :)
func TestSetBody(t *testing.T) {
	fd := map[string]string{
		"foo": "bar baz",
	}

	cases := []struct {
		rawBody      *types.NullableStringOrBytes
		formData     map[string]string
		formEncoding starlark.String
		jsonData     starlark.Value
		body         string
		err          string
	}{
		{types.NewNullableStringOrBytes("hallo"), nil, starlark.String(""), nil, "hallo", ""},
		{types.NewNullableStringOrBytes(""), fd, starlark.String(""), nil, "foo=bar+baz", ""},
		// TODO - this should check multipart form data is being set
		{nil, fd, starlark.String("multipart/form-data"), nil, "", ""},
		{nil, nil, starlark.String(""), starlark.Tuple{starlark.Bool(true), starlark.MakeInt(1), starlark.String("der")}, "[true,1,\"der\"]", ""},
		// the body kinds are mutually exclusive: the old code silently
		// dropped the later ones (or corrupted the request for json+form)
		{types.NewNullableStringOrBytes("hallo"), nil, starlark.String(""), starlark.MakeInt(1), "", "body, json_body and form_body are mutually exclusive, got 2 of them"},
		{types.NewNullableStringOrBytes(""), fd, starlark.String(""), starlark.MakeInt(1), "", "body, json_body and form_body are mutually exclusive, got 2 of them"},
		{types.NewNullableStringOrBytes("hallo"), fd, starlark.String(""), starlark.MakeInt(1), "", "body, json_body and form_body are mutually exclusive, got 3 of them"},
	}

	for i, c := range cases {
		var formData *starlark.Dict
		if c.formData != nil {
			formData = starlark.NewDict(len(c.formData))
			for k, v := range c.formData {
				if err := formData.SetKey(starlark.String(k), starlark.String(v)); err != nil {
					t.Fatal(err)
				}
			}
		}

		req := httptest.NewRequest("get", "https://example.com", nil)
		err := setBody(req, c.rawBody, formData, c.formEncoding, c.jsonData)
		if !(err == nil && c.err == "" || (err != nil && err.Error() == c.err)) {
			t.Errorf("case %d error mismatch. expected: %s, got: %s", i, c.err, err)
			continue
		}

		if strings.HasPrefix(req.Header.Get("Content-Type"), "multipart/form-data;") {
			if err := req.ParseMultipartForm(0); err != nil {
				t.Fatal(err)
			}

			for k, v := range c.formData {
				fv := req.FormValue(k)
				if fv != v {
					t.Errorf("case %d error mismatch. expected %s=%s, got: %s", i, k, v, fv)
				}
			}
		} else {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}

			if string(body) != c.body {
				t.Errorf("case %d body mismatch. expected: %s, got: %s", i, c.body, string(body))
			}
		}
	}
}
