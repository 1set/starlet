package http

import (
	"bytes"
	"net/http"
	"reflect"
	"testing"

	"go.starlark.net/starlark"
)

func TestConvertServerRequest(t *testing.T) {
	// create a request
	jsonBody := []byte(`{"name":"John","age":30}`)
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Errorf("Failed to create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("X-Custom-Header", "Custom Value 1")
	req.Header.Add("X-Custom-Header", "Custom Value 2")
	q := req.URL.Query()
	q.Add("param1", "value1")
	q.Add("param2", "value2")
	q.Add("param2", "value_two")
	req.URL.RawQuery = q.Encode()
	req.RemoteAddr = "127.0.0.1:12345"

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
		{"body", starlark.String(jsonBody)},
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
