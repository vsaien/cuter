package httpx

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/vsaien/cuter/lib/httprouter"
	"github.com/vsaien/cuter/lib/mapping"
)

const (
	multipartFormData = "multipart/form-data"
	xForwardFor       = "X-Forward-For"
	formKey           = "form"
	pathKey           = "path"
	emptyJson         = "{}"
	maxMemory         = 32 << 20 // 32MB
	maxBodyLen        = 1 << 20  // 1MB
	separator         = ";"
	tokensInAttribute = 2
)

var (
	ErrBodylessRequest = errors.New("not a POST|PUT|PATCH request")

	formUnmarshaler = mapping.NewUnmarshaler(formKey, mapping.WithStringValues())
	pathUnmarshaler = mapping.NewUnmarshaler(pathKey, mapping.WithStringValues())
)

type (
	parseFormOptions struct {
		multipartForm bool
	}

	ParseFormOption func(*parseFormOptions)
)

// Returns the peer address, supports X-Forward-For
func GetRemoteAddr(r *http.Request) string {
	v := r.Header.Get(xForwardFor)
	if len(v) > 0 {
		return v
	}
	return r.RemoteAddr
}

func Parse(r *http.Request, v interface{}) error {
	if err := pathPath(r, v); err != nil {
		return err
	}

	if err := parseForm(r, v); err != nil {
		return err
	}

	return parseJsonBody(r, v)
}

func ParseHeader(headerValue string) map[string]string {
	ret := make(map[string]string)
	fields := strings.Split(headerValue, separator)

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if len(field) == 0 {
			continue
		}

		kv := strings.SplitN(field, "=", tokensInAttribute)
		if len(kv) != tokensInAttribute {
			continue
		}

		ret[kv[0]] = kv[1]
	}

	return ret
}

// Parses the form request, supports multipart-form by passing WithMultipartForm()
func parseForm(r *http.Request, v interface{}) error {
	if strings.Index(r.Header.Get(ContentType), multipartFormData) != -1 {
		if err := r.ParseMultipartForm(maxMemory); err != nil {
			return err
		}
	} else {
		if err := r.ParseForm(); err != nil {
			return err
		}
	}

	params := make(map[string]interface{}, len(r.Form))
	for name := range r.Form {
		formValue := r.Form.Get(name)
		if len(formValue) > 0 {
			params[name] = formValue
		}
	}

	return formUnmarshaler.Unmarshal(params, v)
}

// Parses the post request which contains json in body.
func parseJsonBody(r *http.Request, v interface{}) error {
	var reader io.Reader

	if withJsonBody(r) {
		reader = io.LimitReader(r.Body, maxBodyLen)
	} else {
		reader = strings.NewReader(emptyJson)
	}

	return mapping.UnmarshalJsonReader(reader, v)
}

// Parses the symbols reside in url path.
// Like http://localhost/bag/:name
func pathPath(r *http.Request, v interface{}) error {
	vars := httprouter.Vars(r)
	m := make(map[string]interface{}, len(vars))
	for k, v := range vars {
		m[k] = v
	}

	return pathUnmarshaler.Unmarshal(m, v)
}

func withJsonBody(r *http.Request) bool {
	return r.ContentLength > 0 && strings.Index(r.Header.Get(ContentType), ApplicationJson) != -1
}
