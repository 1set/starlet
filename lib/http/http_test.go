package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"

	itn "github.com/1set/starlet/internal"
	lh "github.com/1set/starlet/lib/http"
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
		got, err := lh.AsString(c.in)
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

	thread := &starlark.Thread{Load: itn.NewAssertLoader(lh.ModuleName, lh.LoadModule)}
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
		if r.Header.Get("Task") == "JSON" {
			s := struct {
				Word         string
				ArrayInteger []int
				ArrayDouble  []float64
				Double       float64
				Integer      int
				Bool         bool
				Nothing      interface{}
				Anything     interface{}
				Later        time.Time `json:"then"`
			}{
				Word:         "hello",
				ArrayInteger: []int{1, 2, 3},
				ArrayDouble:  []float64{1.0, 2.1, 3.2},
				Double:       1.2345,
				Integer:      12345,
				Bool:         true,
				Nothing:      nil,
				Anything: map[string]interface{}{
					"foo": "bar",
				},
				Later: time.Date(2023, 7, 15, 9, 30, 0, 0, time.UTC),
			}
			ss, _ := json.Marshal(s)
			w.Write(ss)
		} else {
			w.Write(b)
		}
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
		A: "我爱你",
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
				b = res.body()
				assert.eq(res.status_code, 200)
				assert.true(b.endswith("\r\n\r\n"))
				print(b)
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
				assert.true(res.body().startswith("POST "))
			`),
		},
		{
			name: `POST Default None`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, params=None, headers=None, body=None, json_body=None, form_body=None)
				assert.eq(res.status_code, 200)
				assert.true("null" not in res.body())
				assert.true("None" not in res.body())
			`),
		},
		{
			name: `POST Data String`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, body='Hello')
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.true(b.startswith("POST "))
				assert.true('Content-Length: 5' in b)
				assert.true('Hello' in b)
			`),
		},
		{
			name: `POST Data Bytes`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, body=b'World')
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.true(b.startswith("POST "))
				assert.true('Content-Length: 5' in b)
				assert.true('World' in b)
			`),
		},
		{
			name: `POST JSON String`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, json_body='Hello')
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.true(b.startswith("POST "))
				assert.true('/json' in b)
				assert.true('"Hello"' in b)
			`),
		},
		{
			name: `POST JSON Dict`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, json_body={ "a" : "b", "c" : "d"})
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.true(b.startswith("POST "))
				assert.true('application/json' in b)
				assert.true('{"a":"b","c":"d"}' in b)
			`),
		},
		{
			name: `POST JSON Dict and Params`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, params={"hello": "world"}, json_body={ "a" : "b", "c" : "d"})
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.true(b.startswith("POST "))
				assert.true('application/json' in b)
				assert.true('/?hello=world' in b)
				assert.true('{"a":"b","c":"d"}' in b)
			`),
		},
		{
			name: `POST JSON Dict Number`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, json_body={ 123: 456, 789: 0.123})
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.true(b.startswith("POST "))
				assert.true('/json' in b)
				assert.true('{"123":456,"789":0.123}' in b)
			`),
		},
		{
			name: `POST JSON Struct`,
			script: itn.HereDoc(`
				load('http', 'post')
				load('struct.star', 'struct')
				s = struct(a = 'bee', c = 'dee')
				res = post(test_server_url, json_body=s)
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.true(b.startswith("POST "))
				assert.true('/json' in b)
				assert.true('{"a":"bee","c":"dee"}' in b)
			`),
		},
		{
			name: `POST JSON Module`,
			script: itn.HereDoc(`
				load('http', 'post')
				load('module.star', 'module')
				s = module("data", a = 'egg', c = 'flow')
				res = post(test_server_url, json_body=s)
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.true(b.startswith("POST "))
				assert.true('/json' in b)
				assert.true('{"a":"egg","c":"flow"}' in b)
			`),
		},
		{
			name: `POST JSON Hybrid`,
			script: itn.HereDoc(`
				load('http', 'post')
				load('struct.star', 'struct')
				load('module.star', 'module')
				s = struct(a = 'bee', c = 'dee')
				m = module("data", a = 'egg', c = 'flow')
				d = { "st" : s, "md" : m, 123:456}
				res = post(test_server_url, json_body=d)
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.true(b.startswith("POST "))
				assert.true('/json' in b)
				assert.true('{"123":456,"md":{"a":"egg","c":"flow"},"st":{"a":"bee","c":"dee"}}' in b)
			`),
		},
		{
			name: `POST JSON Starlight`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, json_body=test_custom_data)
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.true(b.startswith("POST "))
				assert.true('/json' in b)
				assert.true('{"A":"我爱你","B":123,"C":true}' in b)
			`),
		},
		{
			name: `POST Form`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, form_body={ "a" : "b", "c" : "d"})
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.true(b.startswith("POST "))
				assert.true('/x-www-form-urlencoded' in b)
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
				assert.true(b.startswith("POST "))
				assert.true('/json' in b)
				assert.true('Foo: bar' in b)
			`),
		},
		{
			name: `POST with UA Set`,
			preset: func() {
				lh.UserAgent = "GqQdYX3eIJw2DTt"
			},
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, json_body={ "a" : "b", "c" : "d"})
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.true(b.startswith("POST "))
				assert.true('/json' in b)
				assert.true('User-Agent: GqQdYX3eIJw2DTt' in b)
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
			name: `Global Timeout`,
			script: itn.HereDoc(`
				load('http', 'get_timeout', 'set_timeout')
				assert.eq(get_timeout(), 30)
				set_timeout(10.5)
				assert.eq(get_timeout(), 10.5)
			`),
		},
		{
			name: `Invalid Timeout`,
			script: itn.HereDoc(`
				load('http', 'get_timeout', 'set_timeout')
				set_timeout("10")
			`),
			wantErr: `http.set_timeout: for parameter timeout: got string, want float or int`,
		},
		{
			name: `Invalid Timeout Value`,
			script: itn.HereDoc(`
				load('http', 'get_timeout', 'set_timeout')
				set_timeout(-10)
			`),
			wantErr: `http.set_timeout: timeout must be non-negative`,
		},
		{
			name: `Invalid Get Timeout`,
			script: itn.HereDoc(`
				load('http', 'get_timeout')
				get_timeout(0.5)
			`),
			wantErr: `http.get_timeout: got 1 arguments, want 0`,
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
			name: `GET Global Timeout`,
			script: itn.HereDoc(`
				load('http', 'get', 'set_timeout')
				set_timeout(0.01)
				res = get(test_server_url)
			`),
			wantErr: `context deadline exceeded (Client.Timeout exceeded while awaiting headers)`,
		},
		{
			name: `GET Not Timeout`,
			script: itn.HereDoc(`
				load('http', 'get', 'set_timeout')
				res = get(test_server_url, timeout=0.5)
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.true(b.startswith("GET /"))
			`),
		},
		{
			name: `GET With No Timeout`,
			script: itn.HereDoc(`
				load('http', 'get')
				res = get(test_server_url, timeout=0)
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.true(b.startswith("GET /"))
			`),
		},
		{
			name: `GET Header`,
			script: itn.HereDoc(`
				load('http', 'get')
				res = get(test_server_url, timeout=0)
				assert.eq(res.status_code, 200)
				head = res.headers
				enc = res.encoding
				assert.eq(head['Content-Type'], 'text/plain; charset=utf-8')
				assert.eq(enc, '')
			`),
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
		{
			name: `POST JSON Marshal`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, headers={ "Task" : "JSON"})
				assert.eq(res.status_code, 200)
				data = res.json()
				print(data)
				[assert.eq(type(i), "int") for i in data['ArrayInteger']]
				assert.eq(type(data['ArrayDouble'][0]), "int")
				assert.eq(type(data['ArrayDouble'][1]), "float")
				assert.eq(type(data['ArrayDouble'][2]), "float")
				assert.eq(type(data['Double']), "float")
				assert.eq(type(data['Integer']), "int")
				assert.eq(type(data['Bool']), "bool")
				assert.true(data['Bool'])
				
				t = data['then']
				assert.eq(type(t), "time.time")
				assert.eq(t.year, 2023)
				assert.eq(t.month, 7)
				assert.eq(t.day, 15)

				text = res.body()
				assert.true('"ArrayInteger":[1,2,3]' in text)
			`),
		},
		{
			name: `POST JSON Marshal Error`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, headers={ "Task" : "JSONError"})
				assert.eq(res.status_code, 200)
				data = res.json()
				assert.eq(data, None)
			`),
		},
		{
			name: `POST Form`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, headers={"ABC": "123"}, form_encoding="multipart/form-data", form_body={ "a" : "better", "c" : "dance", 123: "abc"})
				assert.eq(res.status_code, 200)
				assert.true('POST /' in res.body())
				assert.true('multipart/form-data; boundary=' in res.body())
			`),
		},
		{
			name: `POST postForm`,
			script: itn.HereDoc(`
				load('http', post='postForm')
				res = post(test_server_url, headers={"ABC": "123"}, form_body={ "a" : "better", "c" : "dance"})
				assert.eq(res.status_code, 200)
				assert.true('POST /' in res.body())
				assert.true('application/x-www-form-urlencoded' in res.body())
			`),
		},
		{
			name: `POST Form File`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, headers={"ABC": "123"}, form_body={
					"a" : ["better.txt", "123456"],
					"b" : ["dance.md", '"abcdef(@!'],
				})
				assert.eq(res.status_code, 200)
				rb = res.body()
				assert.true('POST /' in rb)
				assert.true('Content-Type: multipart/form-data; boundary=' in rb)
				assert.true('filename="better.txt"' in rb)
				assert.true('filename="dance.md"' in rb)
				assert.true('Content-Type: application/octet-stream' in rb)
				assert.true('123456' in rb)
				assert.true('"abcdef(@!' in rb)
			`),
		},
		{
			name: `POST Form Key Type`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, headers={"ABC": "123"}, form_body={123: "abc"})
				assert.eq(res.status_code, 200)
				rb = res.body()
				assert.true('POST /' in rb)
				assert.true('Content-Type: application/x-www-form-urlencoded' in rb)
				assert.true('123=abc' in rb)
			`),
		},
		{
			name: `POST Invalid JSON`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, headers={"ABC": "123"}, json_body={
					"fn" : post,
				})
			`),
			wantErr: `unmarshaling starlark value: unrecognized starlark type: *starlark.Builtin`,
		},
		{
			name: `POST Invalid Form Value`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, headers={"ABC": "123"}, form_body={
					"a" : 100,
				})
			`),
			wantErr: `expected param value for key "a" in form_body to be a string or tuple. got: "int"`,
		},
		{
			name: `POST Invalid Form File`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, headers={"ABC": "123"}, form_body={
					"a" : ["better.txt"],
				})
			`),
			wantErr: `expected 2 values for key "a" in form_body to be a tuple of (filename, content)`,
		},
		{
			name: `POST Invalid Form File 1`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, headers={"ABC": "123"}, form_body={
					"a" : [123, "abc"],
				})
			`),
			wantErr: `expected 1st value for key "a" in form_body to be a string. got: "int"`,
		},
		{
			name: `POST Invalid Form File 2`,
			script: itn.HereDoc(`
				load('http', 'post')
				res = post(test_server_url, headers={"ABC": "123"}, form_body={
					"a" : ["abc", [123]],
				})
			`),
			wantErr: `expected 2nd value for key "a" in form_body to be a string. got: "list"`,
		},
		{
			name: `POST Force Form`,
			script: itn.HereDoc(`
				load('http', post='postForm')
				res = post(test_server_url, headers={"ABC": "123"}, form_body={
					"a" : ["better.txt", "123456"],
					"b" : ["dance.md", '"abcdef(@!'],
				})
				assert.eq(res.status_code, 200)
				rb = res.body()
				assert.true('POST /' in rb)
				assert.true('Content-Type: application/x-www-form-urlencoded' in rb)
				assert.true('filename="better.txt"' not in rb)
				assert.true('filename="dance.md"' not in rb)
				assert.true('Content-Type: application/octet-stream' not in rb)
			`),
		},
		{
			name: `Call No Arg`,
			script: itn.HereDoc(`
				load('http', 'call')
				call()
			`),
			wantErr: `http.call: missing method name`,
		},
		{
			name: `Call Invalid Arg`,
			script: itn.HereDoc(`
				load('http', 'call')
				call(123)
			`),
			wantErr: `http.call: for method name: got int, want string or bytes`,
		},
		{
			name: `Call Invalid Method`,
			script: itn.HereDoc(`
				load('http', 'call')
				call("LOVE")
			`),
			wantErr: `unsupported method: love`,
		},
		{
			name: `Simple Call GET`,
			script: itn.HereDoc(`
				load('http', 'call')
				res = call('get', test_server_url)
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.eq(res.status_code, 200)
				assert.true(b.endswith("\r\n\r\n"))
			`),
		},
		{
			name: `Call POST JSON Dict and Params`,
			script: itn.HereDoc(`
				load('http', 'call')
				res = call('POST', test_server_url, params={"hello": "world"}, json_body={ "a" : "b", "c" : "d"})
				assert.eq(res.status_code, 200)
				b = res.body()
				assert.true(b.startswith("POST "))
				assert.true('application/json' in b)
				assert.true('/?hello=world' in b)
				assert.true('{"a":"b","c":"d"}' in b)
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preset != nil {
				tt.preset()
			}
			lh.TimeoutSecond = 30.0
			res, err := itn.ExecModuleWithErrorTest(t, lh.ModuleName, lh.LoadModule, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("http(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}

// DomainWhitelistGuard allows requests only to domains in its whitelist.
type DomainWhitelistGuard struct {
	whitelist map[string]struct{} // Set of allowed domains
}

// NewDomainWhitelistGuard creates a new DomainWhitelistGuard with the specified domains.
func NewDomainWhitelistGuard(domains []string) *DomainWhitelistGuard {
	whitelist := make(map[string]struct{})
	for _, domain := range domains {
		whitelist[domain] = struct{}{}
	}
	return &DomainWhitelistGuard{whitelist: whitelist}
}

// Allowed checks if the request's domain is in the whitelist.
func (g *DomainWhitelistGuard) Allowed(thread *starlark.Thread, req *http.Request) (*http.Request, error) {
	if _, ok := g.whitelist[req.URL.Host]; ok {
		// Domain is in the whitelist, allow the request
		return req, nil
	}
	// Domain is not in the whitelist, deny the request
	return nil, errors.New("request to this domain is not allowed")
}

func TestLoadModule_CustomLoad(t *testing.T) {
	md := lh.NewModule()
	proxyURL, _ := url.Parse("http://127.0.0.1:9999")
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	md.SetClient(client)
	guard := NewDomainWhitelistGuard([]string{"allowed.com"})
	md.SetGuard(guard)

	httpHand := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := httputil.DumpRequest(r, true)
		if err != nil {
			t.Errorf("Error dumping request: %v", err)
		}
		t.Logf("Web server received request: [[%s]]", b)
		time.Sleep(10 * time.Millisecond)
		w.Write(b)
	})
	ts := httptest.NewServer(httpHand)
	defer ts.Close()

	tests := []struct {
		name    string
		preset  func()
		script  string
		wantErr string
	}{
		{
			name: `Simple GET`,
			script: itn.HereDoc(`
				load('http', 'get')
				res = get("http://allowed.com/hello")
				assert.eq(res.status_code, 200)
			`),
			wantErr: `proxyconnect tcp: dial tcp 127.0.0.1:9999: connect`,
		},
		{
			name: `Not Allowed`,
			script: itn.HereDoc(`
				load('http', 'get')
				res = get("http://topsecret.com/text")
				assert.eq(res.status_code, 200)
			`),
			wantErr: `request to this domain is not allowed`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preset != nil {
				tt.preset()
			}
			res, err := itn.ExecModuleWithErrorTest(t, lh.ModuleName, md.LoadModule, tt.script, tt.wantErr, starlark.StringDict{
				"test_server_url": starlark.String(ts.URL),
			})
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("http(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
