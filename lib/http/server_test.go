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
	// Request(body = "{\"name\":\"John\",\"age\":30}", encoding = [], header = {"Content-Type": ["application/json"], "X-Custom-Header": ["Custom Value 1", "Custom Value 2"]}, host = "", json = {"name": "John", "age": 30}, method = "POST", proto = "HTTP/1.1", query = {"param1": ["value1"], "param2": ["value2", "value_two"], "ver": ["0317"]}, remote = "127.0.0.1:12345", url = "/?param1=value1&param2=value2&param2=value_two&ver=0317")
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
		{"body", starlark.String(jsonBody)},
		{"json", sd},
		{"host", starlark.String("")},
		{"remote", starlark.String("127.0.0.1:12345")},
		{"proto", starlark.String("HTTP/1.1")},
		{"method", starlark.String("POST")},
		{"header", sh},
		{"query", sq},
	}
	for _, f := range fields {
		got, err := sr.Attr(f.name)
		if err != nil {
			t.Errorf("fail to get Request(%s) from struct: %v", f.name, err)
			return
		}
		if !reflect.DeepEqual(got, f.want) {
			t.Errorf("Request(%s) got = %v, want %v", f.name, got, f.want)
			return
		}
	}
}
