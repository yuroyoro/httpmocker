package httpmocker

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestMockServer(t *testing.T) {
	drainBody := func(t *testing.T, resp *http.Response) string {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("unexpected error : %+v", err)
		}

		return string(body)
	}

	t.Run("Simple mocking", func(t *testing.T) {
		server := Launch(
			Response{
				Method:      "GET",
				Path:        "/hello",
				Code:        http.StatusOK,
				ContentType: "text/plain",
				Body:        "hello, world",
				Headers:     map[string][]string{"X-Custom-Header": []string{"custom header from mock"}},
			},
		)
		server.Logger = t
		defer server.Close()

		url := fmt.Sprintf("%s/hello", server.URL)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("unexpected error : %+v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status code should be 200 OK : actual %d", resp.StatusCode)
		}

		ctype := resp.Header.Get("Content-Type")
		if ctype != "text/plain" {
			t.Errorf("ContentType should be text/plain: actual %s", ctype)
		}

		xh := resp.Header.Get("X-Custom-Header")
		if xh != "custom header from mock" {
			t.Errorf("X-Custom-Header should be \"custom header from mock\": actual %s", ctype)
		}

		body := drainBody(t, resp)
		if string(body) != "hello, world" {
			t.Errorf("response body should be \"hello, world\": actual %s", string(body))
		}
	})

	t.Run("path and query string", func(t *testing.T) {
		server := Launch(
			Response{
				Method: "GET",
				Path:   "/hello",
				Code:   http.StatusOK,
				Body:   "hello, world",
			},
			Response{
				Method: "GET",
				Path:   "/hello",
				Query:  "dummy=1",
				Code:   http.StatusOK,
				Body:   "hello, world with query string",
			},
		)
		server.Logger = t
		defer server.Close()

		// if no query string is given, mock server should return first mock response
		url := fmt.Sprintf("%s/hello", server.URL)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("unexpected error : %+v", err)
		}

		body := drainBody(t, resp)
		if string(body) != "hello, world" {
			t.Errorf("resonse body should be \"hello, world\": actual %s", string(body))
		}

		// if query string given and matched, mock server should return second mock response
		url = fmt.Sprintf("%s/hello?dummy=1", server.URL)
		resp, err = http.Get(url)
		if err != nil {
			t.Fatalf("unexpected error : %+v", err)
		}

		body = drainBody(t, resp)
		if string(body) != "hello, world with query string" {
			t.Errorf("resonse body should be \"hello, world with query string\": actual %s", string(body))
		}

		// if query string given and not matched, mock server should return first mock response
		url = fmt.Sprintf("%s/hello?dummy=2", server.URL)
		resp, err = http.Get(url)
		if err != nil {
			t.Fatalf("unexpected error : %+v", err)
		}

		body = drainBody(t, resp)
		if string(body) != "hello, world" {
			t.Errorf("resonse body should be \"hello, world\": actual %s", string(body))
		}
	})

	t.Run("with custom handler", func(t *testing.T) {
		server := Launch(
			Response{
				Method: "GET",
				Path:   "/hello",
				Handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					io.WriteString(w, "hello, world from custom handler")
				},
			},
		)
		server.Logger = t
		defer server.Close()

		url := fmt.Sprintf("%s/hello", server.URL)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("unexpected error : %+v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status code should be 200 OK : actual %d", resp.StatusCode)
		}

		body := drainBody(t, resp)
		if string(body) != "hello, world from custom handler" {
			t.Errorf("response body should be \"hello, world from custom handler\": actual %s", string(body))
		}
	})

	t.Run("with unknown request handler", func(t *testing.T) {
		server := Launch()
		server.UnknownRequestHandler = func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, "not found from unknown handler")
		}
		server.Logger = t
		defer server.Close()

		url := fmt.Sprintf("%s/hello", server.URL)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("unexpected error : %+v", err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("status code should be 404 Not Found: actual %d", resp.StatusCode)
		}

		body := drainBody(t, resp)
		if string(body) != "not found from unknown handler" {
			t.Errorf("response body should be \"not found from unknown handler\": actual %s", string(body))
		}
	})

	t.Run("with logger", func(t *testing.T) {
		logger := customLogger{}
		server := Launch()
		server.Add("GET", "/hello", http.StatusOK, "hello, world")
		server.Logger = &logger
		defer server.Close()

		url := fmt.Sprintf("%s/hello", server.URL)
		_, err := http.Get(url)
		if err != nil {
			t.Fatalf("unexpected error : %+v", err)
		}

		if logger.msg != "handler : %s %s -> %+v" {
			t.Errorf("unexpected message is passed to logger : actual : %s", logger.msg)
		}
	})
}

type customLogger struct {
	msg  string
	args []interface{}
}

func (l *customLogger) Logf(msg string, args ...interface{}) {
	l.msg = msg
	l.args = args
}
