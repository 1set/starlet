package http

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	itn "github.com/1set/starlet/lib/internal"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarktest"
)

func TestAsString(t *testing.T) {
	cases := []struct {
		in       starlark.Value
		got, err string
	}{
		{starlark.String("foo"), "foo", ""},
		{starlark.String("\"foo'"), "\"foo'", ""},
		{starlark.Bool(true), "", "invalid syntax"},
	}

	for i, c := range cases {
		got, err := AsString(c.in)
		if !(err == nil && c.err == "" || err != nil && err.Error() == c.err) {
			t.Errorf("case %d error mismatch. expected: '%s', got: '%s'", i, c.err, err)
			continue
		}

		if c.got != got {
			t.Errorf("case %d. expected: '%s', got: '%s'", i, c.got, got)
		}
	}
}

func TestLoadModule_HTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Date", "Mon, 01 Jun 2000 00:00:00 GMT")
		if _, err := w.Write([]byte(`{"hello":"world"}`)); err != nil {
			t.Fatal(err)
		}
	}))
	starlark.Universe["test_server_url"] = starlark.String(ts.URL)

	thread := &starlark.Thread{Load: itn.NewAssertLoader(ModuleName, LoadModule)}
	starlarktest.SetReporter(thread, t)

	code := itn.HereDoc(`
	load('assert.star', 'assert')
	load('http', 'get', 'post')

	res_1 = get(test_server_url, params={ "a" : "b", "c" : "d"})
	assert.eq(res_1.url, test_server_url + "?a=b&c=d")
	assert.eq(res_1.status_code, 200)
	assert.eq(res_1.body(), '{"hello":"world"}')
	assert.eq(res_1.json(), {"hello":"world"})

	assert.eq(res_1.headers, {"Date": "Mon, 01 Jun 2000 00:00:00 GMT", "Content-Length": "17", "Content-Type": "text/plain; charset=utf-8"})

	res_2 = get(test_server_url)
	assert.eq(res_2.json()["hello"], "world")

	headers = {"foo" : "bar"}
	post(test_server_url, json_body={ "a" : "b", "c" : "d"}, headers=headers)
	post(test_server_url, form_body={ "a" : "b", "c" : "d"})
`)

	// Execute test file
	_, err := starlark.ExecFile(thread, "test.star", code, nil)
	if err != nil {
		t.Error(err)
	}
}

// we're ok with testing private functions if it simplifies the test :)
func TestSetBody(t *testing.T) {
	fd := map[string]string{
		"foo": "bar baz",
	}

	cases := []struct {
		rawBody      starlark.String
		formData     map[string]string
		formEncoding starlark.String
		jsonData     starlark.Value
		body         string
		err          string
	}{
		{starlark.String("hallo"), nil, starlark.String(""), nil, "hallo", ""},
		{starlark.String(""), fd, starlark.String(""), nil, "foo=bar+baz", ""},
		// TODO - this should check multipart form data is being set
		{starlark.String(""), fd, starlark.String("multipart/form-data"), nil, "", ""},
		{starlark.String(""), nil, starlark.String(""), starlark.Tuple{starlark.Bool(true), starlark.MakeInt(1), starlark.String("der")}, "[true,1,\"der\"]", ""},
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
