package http

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"
	"time"

	itn "github.com/1set/starlet/lib/internal"
	"github.com/1set/starlight/convert"
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

func TestLoadModule_HTTP_One(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Date", "Mon, 01 Jun 2000 00:00:00 GMT")
		if _, err := w.Write([]byte(`{"hello":"world"}`)); err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()
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

func TestLoadModule_HTTP(t *testing.T) {
	httpHand := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := httputil.DumpRequest(r, true)
		if err != nil {
			t.Errorf("Error dumping request: %v", err)
		}
		t.Logf("Web server received request: [[%s]]", b)
		time.Sleep(50 * time.Millisecond)
		w.Write(b)
	})
	ts := httptest.NewServer(httpHand)
	defer ts.Close()
	starlark.Universe["test_server_url"] = starlark.String(ts.URL)

	ts2 := httptest.NewTLSServer(httpHand)
	defer ts2.Close()
	starlark.Universe["test_server_url_ssl"] = starlark.String(ts2.URL)

	starlark.Universe["test_custom_data"] = convert.NewStruct(struct {
		A string
		B int
		C bool
	}{
		A: "foo",
		B: 123,
		C: true,
	})

	tests := []struct {
		name    string
		preset  func()
		script  string
		wantErr string
	}{
		{
			name: `Invalid URL`,
			script: itn.HereDoc(`
				load('http', 'get')
				res = get(123)
			`),
			wantErr: `http.get: for parameter url: got int, want string`,
		},
		{
			name: `Simple GET`,
			script: itn.HereDoc(`
				load('http', 'get')
				res = get(test_server_url)
				assert.eq(res.status_code, 200)
				print(res.body())
			`),
		},
		{
			name: `GET with params`,
			script: itn.HereDoc(`
				load('http', 'get')
				res = get(test_server_url, params={ "a" : "b", "c" : "d"})
				assert.eq(res.url, test_server_url + "?a=b&c=d")
				assert.eq(res.status_code, 200)
			`),
		},
		{
			name: `Simple POST`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url)
				assert.eq(res.status_code, 200)
				assert.eq(res.body().startswith("POST "), True)
			`),
		},
		{
			name: `POST JSON Native`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, json_body={ "a" : "b", "c" : "d"})
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.eq(b.startswith("POST "), True)
				assert.eq('/json' in b, True)
				assert.eq('{"a":"b","c":"d"}' in b, True)
			`),
		},
		{
			name: `POST JSON Converted`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, json_body=test_custom_data)
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.eq(b.startswith("POST "), True)
				assert.eq('/json' in b, True)
				assert.eq('{"A":"foo","B":123,"C":true}' in b, True)
			`),
		},
		{
			name: `POST Form`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, form_body={ "a" : "b", "c" : "d"})
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.eq(b.startswith("POST "), True)
				assert.eq('/x-www-form-urlencoded' in b, True)
			`),
		},
		{
			name: `POST with headers`,
			script: itn.HereDoc(`
				load('http', 'post')
				headers = {"foo" : "bar"}
				res = post(test_server_url, json_body={ "a" : "b", "c" : "d"}, headers=headers)
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.eq(b.startswith("POST "), True)
				assert.eq('/json' in b, True)
				assert.eq('Foo: bar' in b, True)
			`),
		},
		{
			name: `POST with UA Set`,
			preset: func() {
				UserAgent = "GqQdYX3eIJw2DTt"
			},
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, json_body={ "a" : "b", "c" : "d"})
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.eq(b.startswith("POST "), True)
				assert.eq('/json' in b, True)
				assert.eq('User-Agent: GqQdYX3eIJw2DTt' in b, True)
			`),
		},
		{
			name: `HEAD Anything`,
			script: itn.HereDoc(`
				load('http', 'head')
				res = head(test_server_url, params={ "a" : "b", "c" : "d"}, headers={ "foo" : "bar"})
				assert.eq(res.status_code, 200)
				assert.eq(len(res.body()), 0)
			`),
		},
		{
			name: `GET Timeout`,
			script: itn.HereDoc(`
				load('http', 'get')
				res = get(test_server_url, timeout=0.01)
			`),
			wantErr: `context deadline exceeded (Client.Timeout exceeded while awaiting headers)`,
		},
		{
			name: `GET SSL Error`,
			script: itn.HereDoc(`
				load('http', 'get')
				res = get(test_server_url_ssl)
			`),
			wantErr: `x509: certificate signed by unknown authority`,
		},
		{
			name: `GET SSL Insecure`,
			script: itn.HereDoc(`
				load('http', 'get')
				res = get(test_server_url_ssl, verify=False)
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preset != nil {
				tt.preset()
			}
			res, err := itn.ExecModuleWithErrorTest(t, ModuleName, LoadModule, tt.script, tt.wantErr)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("http(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
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
