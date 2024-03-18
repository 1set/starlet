package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"reflect"
	"strings"
	"testing"

	itn "github.com/1set/starlet/internal"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarktest"
	"go.starlark.net/syntax"
)

func TestConvertServerRequest(t *testing.T) {
	// just a taste
	if r := ConvertServerRequest(nil); r != nil {
		t.Errorf("ConvertServerRequest(nil) = %v, want nil", r)
		return
	}

	// create a request
	s := `{"name":"John","age":30}`
	req := getMockRequest(s)

	// do the convert
	sr := ConvertServerRequest(req)

	// check the result
	if sr == nil {
		t.Error("ConvertServerRequest returned nil")
		return
	}

	// prepare expected values
	sd := starlark.NewDict(2)
	_ = sd.SetKey(starlark.String("name"), starlark.String("John"))
	_ = sd.SetKey(starlark.String("age"), starlark.MakeInt(30))
	sh := mapStrs2Dict(map[string][]string{
		"Content-Type":    {"application/json"},
		"X-Custom-Header": {"Custom Value 1", "Custom Value 2"},
	})
	sq := mapStrs2Dict(map[string][]string{
		"param1": {"value1"},
		"param2": {"value2", "value_two"},
	})
	fields := []struct {
		name string
		want starlark.Value
	}{
		{"host", starlark.String("")},
		{"remote", starlark.String("127.0.0.1:12345")},
		{"proto", starlark.String("HTTP/1.1")},
		{"method", starlark.String("POST")},
		{"body", starlark.String(s)},
		{"json", sd},
		{"header", sh},
		{"query", sq},
	}

	// check the fields
	for _, f := range fields {
		got, err := sr.Attr(f.name)
		if err != nil {
			t.Errorf("fail to get Request(%s) from struct: %v", f.name, err)
			return
		}
		if yes, err := starlark.Equal(got, f.want); err == nil {
			if !yes {
				t.Errorf("Request(%s)[SL] got = %v, want %v", f.name, got, f.want)
				return
			}
		} else if !reflect.DeepEqual(got, f.want) {
			t.Errorf("Request(%s)[GO] got = %v, want %v", f.name, got, f.want)
			return
		}
	}
}

func TestServerResponse(t *testing.T) {
	if sr := NewServerResponse(); sr == nil {
		t.Error("NewServerResponse returned nil")
		return
	} else if err := sr.Write(nil); err == nil {
		t.Error("ServerResponse.Write(nil) returned nil")
		return
	}

	bd := `{"name":"John","age":30}`
	testCases := []struct {
		name             string
		script           string
		request          *http.Request
		expectedStatus   int
		expectedResponse string
	}{
		{
			name: "no ops",
			script: itn.HereDoc(`
			`),
			request:        getMockRequest(bd),
			expectedStatus: http.StatusOK,
			expectedResponse: itn.HereDoc(`
				Content-Type: application/octet-stream
			`),
		},
		{
			name: "full json and override",
			script: itn.HereDoc(`
				print(request)
				response.set_status(201)
				response.add_header("Content-Type", "Not Seen")
				response.add_header("X-Think", "Testing")
				response.add_header("X-Think", "Starlark")
				response.set_content_type("")
				response.set_data(b'Not Data')
				response.set_text(b'Hello, World!')
				response.set_html('<h1>Hello, World!</h1>')
				response.set_json({"abc": [1, 2, 3]})
			`),
			request:        getMockRequest(bd),
			expectedStatus: http.StatusCreated,
			expectedResponse: itn.HereDoc(`
				Content-Type: application/json
				X-Think: Testing
				X-Think: Starlark
				{"abc":[1,2,3]}
			`),
		},
		{
			name: "simple text",
			script: itn.HereDoc(`
				response.set_text('Hello, World!')
			`),
			request:        getMockRequest(bd),
			expectedStatus: http.StatusOK,
			expectedResponse: itn.HereDoc(`
				Content-Type: text/plain
				Hello, World!
			`),
		},
		{
			name: "simple html",
			script: itn.HereDoc(`
				response.set_html('<h1>Hello, World!</h1>')
			`),
			request:        getMockRequest(bd),
			expectedStatus: http.StatusOK,
			expectedResponse: itn.HereDoc(`
				Content-Type: text/html
				<h1>Hello, World!</h1>
			`),
		},
		{
			name: "set content type",
			script: itn.HereDoc(`
				response.set_content_type("application/starlark")
			`),
			request:        getMockRequest(bd),
			expectedStatus: http.StatusOK,
			expectedResponse: itn.HereDoc(`
				Content-Type: application/starlark
			`),
		},
		{
			name: "invalid status code",
			script: itn.HereDoc(`
				response.set_code(999)
			`),
			request:        getMockRequest(bd),
			expectedStatus: http.StatusBadRequest,
			expectedResponse: itn.HereDoc(`
		        Content-Type: text/plain
		        invalid status code
			`),
		},
		{
			name: "status code 101",
			script: itn.HereDoc(`
				response.set_code(101)
				response.set_text('Switch')
			`),
			request:        getMockRequest(bd),
			expectedStatus: http.StatusSwitchingProtocols,
			expectedResponse: itn.HereDoc(`
		        Content-Type: text/plain
		        Switch
			`),
		},
		{
			name: "invalid json",
			script: itn.HereDoc(`
				d = {"abc": [1, 2, 3]}
				d["circular"] = d
				response.set_json(d)
			`),
			request:        getMockRequest(bd),
			expectedStatus: http.StatusBadRequest,
			expectedResponse: itn.HereDoc(`
		        Content-Type: text/plain
				cyclic reference found
			`),
		},
		{
			name: "invalid json in request",
			script: itn.HereDoc(`
				response.set_data(str(request.json == None))
			`),
			request:        getMockRequest(`{"name":"John","age":30`),
			expectedStatus: http.StatusOK,
			expectedResponse: itn.HereDoc(`
				Content-Type: application/octet-stream
				True
			`),
		},
	}
	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// call the handler
			rr := httptest.NewRecorder()
			handler := getScriptHandler(tc.script)
			handler(rr, tc.request)

			// check the status
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Unexpected status code: %v, expected %v", status, tc.expectedStatus)
			}

			// check the response by comparing the expected lines
			if bs, err := httputil.DumpResponse(rr.Result(), true); err != nil {
				t.Errorf("fail to dump response: %v", err)
			} else {
				t.Logf("Response#%d: %s", i, bs)
				expLines := strings.Split(strings.TrimSpace(tc.expectedResponse), "\n")
				got := string(bs)
				gotLines := strings.Split(strings.TrimSpace(got), "\n")
				for _, expLine := range expLines {
					if expLine = strings.TrimSpace(expLine); expLine == "" {
						continue
					}
					found := false
					for _, gotLine := range gotLines {
						if gotLine = strings.TrimSpace(gotLine); gotLine == "" {
							continue
						}
						if strings.Contains(gotLine, expLine) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected line not found in response: {%v}", expLine)
					}
				}
			}
		})
	}
}

func getMockRequest(s string) *http.Request {
	jsonBody := []byte(s)
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("X-Custom-Header", "Custom Value 1")
	req.Header.Add("X-Custom-Header", "Custom Value 2")
	q := req.URL.Query()
	q.Add("param1", "value1")
	q.Add("param2", "value2")
	q.Add("param2", "value_two")
	req.URL.RawQuery = q.Encode()
	req.RemoteAddr = "127.0.0.1:12345"
	return req
}

func getScriptHandler(script string) func(w http.ResponseWriter, r *http.Request) {
	// create a new http handler
	return func(w http.ResponseWriter, r *http.Request) {
		// prepare envs
		resp := NewServerResponse()
		pred := starlark.StringDict{
			"request":  ConvertServerRequest(r),
			"response": resp.Struct(),
		}

		// execute the script
		thread := &starlark.Thread{Name: "http"}
		starlarktest.SetReporter(thread, nil)
		opts := syntax.FileOptions{
			Set:            true,
			GlobalReassign: true,
		}
		_, err := starlark.ExecFileOptions(&opts, thread, "handler.star", []byte(script), pred)

		// handle error
		if err != nil {
			w.Header().Add("Content-Type", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// handle response
		if err := resp.Write(w); err != nil {
			w.Header().Add("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
		return
	}
}
