package http

import (
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/1set/starlet/dataconv/types"
	"go.starlark.net/starlark"
)

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
