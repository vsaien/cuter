package httpx

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/vsaien/cuter/lib/httprouter"

	"github.com/stretchr/testify/assert"
)

const (
	applicationJsonWithUtf8 = "application/json; charset=utf-8"
	contentLength           = "Content-Length"
)

func TestGetRemoteAddr(t *testing.T) {
	host := "8.8.8.8"
	r, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	assert.Nil(t, err)

	r.Header.Set(xForwardFor, host)
	assert.Equal(t, host, GetRemoteAddr(r))
}

func TestParseForm(t *testing.T) {
	v := struct {
		Name    string  `form:"name"`
		Age     int     `form:"age"`
		Percent float64 `form:"percent,optional"`
	}{}

	r, err := http.NewRequest(http.MethodGet, "http://hello.com/a?name=hello&age=18&percent=3.4", nil)
	assert.Nil(t, err)

	err = Parse(r, &v)
	assert.Nil(t, err)
	assert.Equal(t, "hello", v.Name)
	assert.Equal(t, 18, v.Age)
	assert.Equal(t, 3.4, v.Percent)
}

func TestParseRequired(t *testing.T) {
	v := struct {
		Name    string  `form:"name"`
		Percent float64 `form:"percent"`
	}{}

	r, err := http.NewRequest(http.MethodGet, "http://hello.com/a?name=hello", nil)
	assert.Nil(t, err)

	err = Parse(r, &v)
	assert.NotNil(t, err)
}

func TestParseSlice(t *testing.T) {
	body := `names=%5B%22first%22%2C%22second%22%5D`
	reader := strings.NewReader(body)
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/", reader)
	assert.Nil(t, err)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Names []string `form:"names"`
		}{}

		err = Parse(r, &v)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(v.Names))
		assert.Equal(t, "first", v.Names[0])
		assert.Equal(t, "second", v.Names[1])
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
}

func TestParseJsonPost(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/kevin/2017?nickname=whatever&zipcode=200000",
		bytes.NewBufferString(`{"location": "shanghai", "time": 20170912}`))
	r.Header.Set(ContentType, ApplicationJson)
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name     string `path:"name"`
			Year     int    `path:"year"`
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode"`
			Location string `json:"location"`
			Time     int64  `json:"time"`
		}{}

		err = Parse(r, &v)
		assert.Nil(t, err)
		io.WriteString(w, fmt.Sprintf("%s:%d:%s:%d:%s:%d", v.Name, v.Year,
			v.Nickname, v.Zipcode, v.Location, v.Time))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)

	assert.Equal(t, "kevin:2017:whatever:200000:shanghai:20170912", rr.Body.String())
}

func TestParseJsonPostError(t *testing.T) {
	payload := `[{"abcd": "cdef"}]`
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/kevin/2017?nickname=whatever&zipcode=200000",
		bytes.NewBufferString(payload))
	r.Header.Set(ContentType, ApplicationJson)
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name     string `path:"name"`
			Year     int    `path:"year"`
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode"`
			Location string `json:"location"`
			Time     int64  `json:"time"`
		}{}

		err = Parse(r, &v)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), payload))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
}

func TestParseJsonPostRequired(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/kevin/2017",
		bytes.NewBufferString(`{"location": "shanghai"`))
	r.Header.Set(ContentType, ApplicationJson)
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name     string `path:"name"`
			Year     int    `path:"year"`
			Location string `json:"location"`
			Time     int64  `json:"time"`
		}{}

		err = Parse(r, &v)
		assert.NotNil(t, err)
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
}

func TestParsePath(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://hello.com/kevin/2017", nil)
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodGet, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name string `path:"name"`
			Year int    `path:"year"`
		}{}

		err = Parse(r, &v)
		assert.Nil(t, err)
		io.WriteString(w, fmt.Sprintf("%s in %d", v.Name, v.Year))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)

	assert.Equal(t, "kevin in 2017", rr.Body.String())
}

func TestParsePathRequired(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://hello.com/kevin", nil)
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodGet, "/:name/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name string `path:"name"`
			Year int    `path:"year"`
		}{}

		err = Parse(r, &v)
		assert.NotNil(t, err)
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
}

func TestParseQuery(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://hello.com/kevin/2017?nickname=whatever&zipcode=200000", nil)
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodGet, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode"`
		}{}

		err = Parse(r, &v)
		assert.Nil(t, err)
		io.WriteString(w, fmt.Sprintf("%s:%d", v.Nickname, v.Zipcode))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)

	assert.Equal(t, "whatever:200000", rr.Body.String())
}

func TestParseQueryRequired(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/kevin/2017?nickname=whatever", nil)
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode"`
		}{}

		err = Parse(r, &v)
		assert.NotNil(t, err)
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
}

func TestParseOptional(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://hello.com/kevin/2017?nickname=whatever&zipcode=", nil)
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodGet, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode,optional"`
		}{}

		err = Parse(r, &v)
		assert.Nil(t, err)
		io.WriteString(w, fmt.Sprintf("%s:%d", v.Nickname, v.Zipcode))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)

	assert.Equal(t, "whatever:0", rr.Body.String())
}

func TestParseNestedInRequestEmpty(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/kevin/2017", bytes.NewBufferString("{}"))
	assert.Nil(t, err)

	type (
		Request struct {
			Name string `path:"name"`
			Year int    `path:"year"`
		}

		Audio struct {
			Volume int `json:"volume"`
		}

		WrappedRequest struct {
			Request
			Audio Audio `json:"audio,optional"`
		}
	)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v WrappedRequest
		err = Parse(r, &v)
		assert.Nil(t, err)
		io.WriteString(w, fmt.Sprintf("%s:%d", v.Name, v.Year))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)

	assert.Equal(t, "kevin:2017", rr.Body.String())
}

func TestParsePtrInRequest(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/kevin/2017",
		bytes.NewBufferString(`{"audio": {"volume": 100}}`))
	r.Header.Set(ContentType, ApplicationJson)
	assert.Nil(t, err)

	type (
		Request struct {
			Name string `path:"name"`
			Year int    `path:"year"`
		}

		Audio struct {
			Volume int `json:"volume"`
		}

		WrappedRequest struct {
			Request
			Audio *Audio `json:"audio,optional"`
		}
	)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v WrappedRequest
		err = Parse(r, &v)
		assert.Nil(t, err)
		io.WriteString(w, fmt.Sprintf("%s:%d:%d", v.Name, v.Year, v.Audio.Volume))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)

	assert.Equal(t, "kevin:2017:100", rr.Body.String())
}

func TestParsePtrInRequestEmpty(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/kevin", bytes.NewBufferString("{}"))
	assert.Nil(t, err)

	type (
		Audio struct {
			Volume int `json:"volume"`
		}

		WrappedRequest struct {
			Audio *Audio `json:"audio,optional"`
		}
	)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/kevin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v WrappedRequest
		err = Parse(r, &v)
		assert.Nil(t, err)
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
}

func TestParseQueryOptional(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://hello.com/kevin/2017?nickname=whatever&zipcode=", nil)
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodGet, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode,optional"`
		}{}

		err = Parse(r, &v)
		assert.Nil(t, err)
		io.WriteString(w, fmt.Sprintf("%s:%d", v.Nickname, v.Zipcode))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)

	assert.Equal(t, "whatever:0", rr.Body.String())
}

func TestParse(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://hello.com/kevin/2017?nickname=whatever&zipcode=200000", nil)
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodGet, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name     string `path:"name"`
			Year     int    `path:"year"`
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode"`
		}{}

		err = Parse(r, &v)
		assert.Nil(t, err)
		io.WriteString(w, fmt.Sprintf("%s:%d:%s:%d", v.Name, v.Year, v.Nickname, v.Zipcode))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)

	assert.Equal(t, "kevin:2017:whatever:200000", rr.Body.String())
}

func TestParseWrappedRequest(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://hello.com/kevin/2017", nil)
	assert.Nil(t, err)

	type (
		Request struct {
			Name string `path:"name"`
			Year int    `path:"year"`
		}

		WrappedRequest struct {
			Request
		}
	)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodGet, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v WrappedRequest
		err = Parse(r, &v)
		assert.Nil(t, err)
		io.WriteString(w, fmt.Sprintf("%s:%d", v.Name, v.Year))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)

	assert.Equal(t, "kevin:2017", rr.Body.String())
}

func TestParseWrappedGetRequestWithJsonHeader(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://hello.com/kevin/2017", nil)
	r.Header.Set(ContentType, applicationJsonWithUtf8)
	assert.Nil(t, err)

	type (
		Request struct {
			Name string `path:"name"`
			Year int    `path:"year"`
		}

		WrappedRequest struct {
			Request
		}
	)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodGet, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v WrappedRequest
		err = Parse(r, &v)
		assert.Nil(t, err)
		io.WriteString(w, fmt.Sprintf("%s:%d", v.Name, v.Year))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)

	assert.Equal(t, "kevin:2017", rr.Body.String())
}

func TestParseWrappedHeadRequestWithJsonHeader(t *testing.T) {
	r, err := http.NewRequest(http.MethodHead, "http://hello.com/kevin/2017", nil)
	r.Header.Set(ContentType, applicationJsonWithUtf8)
	assert.Nil(t, err)

	type (
		Request struct {
			Name string `path:"name"`
			Year int    `path:"year"`
		}

		WrappedRequest struct {
			Request
		}
	)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodHead, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v WrappedRequest
		err = Parse(r, &v)
		assert.Nil(t, err)
		io.WriteString(w, fmt.Sprintf("%s:%d", v.Name, v.Year))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)

	assert.Equal(t, "kevin:2017", rr.Body.String())
}

func TestParseWrappedRequestPtr(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://hello.com/kevin/2017", nil)
	assert.Nil(t, err)

	type (
		Request struct {
			Name string `path:"name"`
			Year int    `path:"year"`
		}

		WrappedRequest struct {
			*Request
		}
	)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodGet, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v WrappedRequest
		err = Parse(r, &v)
		assert.Nil(t, err)
		io.WriteString(w, fmt.Sprintf("%s:%d", v.Name, v.Year))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)

	assert.Equal(t, "kevin:2017", rr.Body.String())
}

func TestParseWithAll(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/kevin/2017?nickname=whatever&zipcode=200000",
		bytes.NewBufferString(`{"location": "shanghai", "time": 20170912}`))
	r.Header.Set(ContentType, ApplicationJson)
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name     string `path:"name"`
			Year     int    `path:"year"`
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode"`
			Location string `json:"location"`
			Time     int64  `json:"time"`
		}{}

		err = Parse(r, &v)
		assert.Nil(t, err)
		io.WriteString(w, fmt.Sprintf("%s:%d:%s:%d:%s:%d", v.Name, v.Year,
			v.Nickname, v.Zipcode, v.Location, v.Time))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)

	assert.Equal(t, "kevin:2017:whatever:200000:shanghai:20170912", rr.Body.String())
}

func TestParseWithAllUtf8(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/kevin/2017?nickname=whatever&zipcode=200000",
		bytes.NewBufferString(`{"location": "shanghai", "time": 20170912}`))
	r.Header.Set(ContentType, applicationJsonWithUtf8)
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name     string `path:"name"`
			Year     int    `path:"year"`
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode"`
			Location string `json:"location"`
			Time     int64  `json:"time"`
		}{}

		err = Parse(r, &v)
		assert.Nil(t, err)
		io.WriteString(w, fmt.Sprintf("%s:%d:%s:%d:%s:%d", v.Name, v.Year,
			v.Nickname, v.Zipcode, v.Location, v.Time))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)

	assert.Equal(t, "kevin:2017:whatever:200000:shanghai:20170912", rr.Body.String())
}

func TestParseWithMissingForm(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/kevin/2017?nickname=whatever",
		bytes.NewBufferString(`{"location": "shanghai", "time": 20170912}`))
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name     string `path:"name"`
			Year     int    `path:"year"`
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode"`
			Location string `json:"location"`
			Time     int64  `json:"time"`
		}{}

		err = Parse(r, &v)
		assert.NotNil(t, err)
		assert.Equal(t, "field zipcode is not set", err.Error())
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
}

func TestParseWithMissingAllForms(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/kevin/2017",
		bytes.NewBufferString(`{"location": "shanghai", "time": 20170912}`))
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name     string `path:"name"`
			Year     int    `path:"year"`
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode"`
			Location string `json:"location"`
			Time     int64  `json:"time"`
		}{}

		err = Parse(r, &v)
		assert.NotNil(t, err)
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
}

func TestParseWithMissingJson(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/kevin/2017?nickname=whatever&zipcode=200000",
		bytes.NewBufferString(`{"location": "shanghai"}`))
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name     string `path:"name"`
			Year     int    `path:"year"`
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode"`
			Location string `json:"location"`
			Time     int64  `json:"time"`
		}{}

		err = Parse(r, &v)
		assert.NotEqual(t, io.EOF, err)
		assert.NotNil(t, Parse(r, &v))
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
}

func TestParseWithMissingAllJsons(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://hello.com/kevin/2017?nickname=whatever&zipcode=200000", nil)
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodGet, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name     string `path:"name"`
			Year     int    `path:"year"`
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode"`
			Location string `json:"location"`
			Time     int64  `json:"time"`
		}{}

		err = Parse(r, &v)
		assert.NotEqual(t, io.EOF, err)
		assert.NotNil(t, err)
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
}

func TestParseWithMissingPath(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/2017?nickname=whatever&zipcode=200000",
		bytes.NewBufferString(`{"location": "shanghai", "time": 20170912}`))
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name     string `path:"name"`
			Year     int    `path:"year"`
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode"`
			Location string `json:"location"`
			Time     int64  `json:"time"`
		}{}

		err = Parse(r, &v)
		assert.NotNil(t, err)
		assert.Equal(t, "field name is not set", err.Error())
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
}

func TestParseWithMissingAllPaths(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/?nickname=whatever&zipcode=200000",
		bytes.NewBufferString(`{"location": "shanghai", "time": 20170912}`))
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name     string `path:"name"`
			Year     int    `path:"year"`
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode"`
			Location string `json:"location"`
			Time     int64  `json:"time"`
		}{}

		err = Parse(r, &v)
		assert.NotNil(t, err)
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
}

func TestParseGetWithContentLengthHeader(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://hello.com/kevin/2017?nickname=whatever&zipcode=200000", nil)
	r.Header.Set(ContentType, ApplicationJson)
	r.Header.Set(contentLength, "1024")
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodGet, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name     string `path:"name"`
			Year     int    `path:"year"`
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode"`
			Location string `json:"location"`
			Time     int64  `json:"time"`
		}{}

		err = Parse(r, &v)
		assert.NotNil(t, err)
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
}

func TestParseJsonPostWithTypeMismatch(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/kevin/2017?nickname=whatever&zipcode=200000",
		bytes.NewBufferString(`{"time": "20170912"}`))
	r.Header.Set(ContentType, applicationJsonWithUtf8)
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name     string `path:"name"`
			Year     int    `path:"year"`
			Nickname string `form:"nickname"`
			Zipcode  int64  `form:"zipcode"`
			Time     int64  `json:"time"`
		}{}

		err = Parse(r, &v)
		assert.NotNil(t, err)
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
}

func TestParseJsonPostWithInt2String(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://hello.com/kevin/2017",
		bytes.NewBufferString(`{"time": 20170912}`))
	r.Header.Set(ContentType, applicationJsonWithUtf8)
	assert.Nil(t, err)

	router := httprouter.NewPatRouter()
	router.Handle(http.MethodPost, "/:name/:year", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := struct {
			Name string `path:"name"`
			Year int    `path:"year"`
			Time string `json:"time"`
		}{}

		err = Parse(r, &v)
		assert.NotNil(t, err)
	}))

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, r)
}

func BenchmarkParseRaw(b *testing.B) {
	r, err := http.NewRequest(http.MethodGet, "http://hello.com/a?name=hello&age=18&percent=3.4", nil)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		v := struct {
			Name    string  `form:"name"`
			Age     int     `form:"age"`
			Percent float64 `form:"percent,optional"`
		}{}

		v.Name = r.FormValue("name")
		v.Age, err = strconv.Atoi(r.FormValue("age"))
		if err != nil {
			b.Fatal(err)
		}
		v.Percent, err = strconv.ParseFloat(r.FormValue("percent"), 64)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseAuto(b *testing.B) {
	r, err := http.NewRequest(http.MethodGet, "http://hello.com/a?name=hello&age=18&percent=3.4", nil)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		v := struct {
			Name    string  `form:"name"`
			Age     int     `form:"age"`
			Percent float64 `form:"percent,optional"`
		}{}

		if err = Parse(r, &v); err != nil {
			b.Fatal(err)
		}
	}
}
