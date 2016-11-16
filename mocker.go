package httpmocker

import (
	"io"
	"net/http"
	"net/http/httptest"
)

// Server : mock server object
type Server struct {
	Server    *httptest.Server
	Responses map[string]map[string][]*Response
	URL       string
	Logger
	UnknownRequestHandler http.HandlerFunc
}

// Response : mocke response
type Response struct {
	Method      string
	Path        string
	Query       string
	Code        int
	ContentType string
	Body        string
	Headers     http.Header

	Handler http.HandlerFunc
}

// Logger : logger for mock server
type Logger interface {
	Logf(string, ...interface{})
}

// Close : shutdown mock server
func (server *Server) Close() {
	if server.Server != nil {
		server.Server.Close()
	}
}

// Add : add mock response to mock server
func (server *Server) Add(method, path string, code int, body string) *Server {
	server.AddResponses(Response{
		Method: method,
		Path:   path,
		Code:   code,
		Body:   body,
	})

	return server
}

// AddEmptyResponse : add empyt mock response to mock server
func (server *Server) AddEmptyResponse(method, path string, code int) *Server {
	server.AddResponses(Response{
		Method: method,
		Path:   path,
		Code:   code,
	})

	return server
}

// AddResponses : add mock response to mock server
func (server *Server) AddResponses(responses ...Response) *Server {

	for _, response := range responses {
		r := response
		m := server.Responses[r.Method]
		if m == nil {
			m = map[string][]*Response{}
			server.Responses[r.Method] = m
		}
		resps := m[r.Path]
		if resps == nil {
			m[r.Path] = []*Response{&r}
			continue
		}

		m[r.Path] = append(m[r.Path], &r)
	}

	return server
}

func (server *Server) findResponse(r *http.Request) *Response {
	method := r.Method
	path := r.URL.Path

	m := server.Responses[method]
	if m == nil {
		return nil
	}

	resps := m[path]
	if len(resps) == 0 {
		return nil
	}

	var candidate *Response
	for _, resp := range resps {
		if resp.Path == path {
			if resp.Query == "" {
				candidate = resp
			}

			if resp.Query != "" && resp.Query == r.URL.RawQuery {
				return resp
			}
		}
	}

	return candidate
}

func (server *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path

	resp := server.findResponse(r)

	// not found
	if resp == nil {
		server.logf("unknown request: %s %s", method, path)
		if server.UnknownRequestHandler != nil {
			server.UnknownRequestHandler(w, r)
		}

		return
	}

	// Send response.

	if resp.Handler != nil {
		// if Handler is set, delegate response
		resp.Handler(w, r)
		return
	}

	header := w.Header()
	header.Set("Content-Type", resp.ContentType)
	if resp.Headers != nil {
		for k := range resp.Headers {
			v := resp.Headers.Get(k)
			header.Set(k, v)
		}
	}
	if resp.Code != 0 {
		w.WriteHeader(resp.Code)
	}

	io.WriteString(w, resp.Body)

	server.logf("handler : %s %s -> %+v", method, path, resp)
	return
}

func (server *Server) logf(msg string, args ...interface{}) {
	if server.Logger != nil {
		server.Logger.Logf(msg, args...)
	}
}

// Start : start up mock server
func (server *Server) Start() *Server {
	httptestserver := httptest.NewServer(
		http.HandlerFunc(server.handleRequest),
	)
	server.Server = httptestserver
	server.URL = httptestserver.URL
	return server
}

// Launch : launch mock server with given mock requests
func Launch(responses ...Response) *Server {
	server := Server{}
	server.Responses = map[string]map[string][]*Response{}
	server.AddResponses(responses...)
	server.Start()

	return &server
}
