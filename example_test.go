package httpmocker_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/yuroyoro/httpmocker"
)

func ExampleHTTPMocker() {
	server := httpmocker.Launch().Add(
		"GET",
		"/hello",
		http.StatusOK,
		"hello, world",
	)
	defer server.Close()

	url := fmt.Sprintf("%s/hello", server.URL)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}

	fmt.Println(resp.Status)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}
	fmt.Println(string(body))

	// Output:
	// 200 OK
	// hello, world
}

func ExampleHTTPMockerWithHeader() {
	server := httpmocker.Launch(
		httpmocker.Response{
			Method:      "GET",
			Path:        "/hello",
			Code:        http.StatusOK,
			ContentType: "text/plain",
			Body:        "hello, world",
			Headers:     map[string][]string{"X-Custom-Header": []string{"custom header from mock"}},
		},
	)
	defer server.Close()

	url := fmt.Sprintf("%s/hello", server.URL)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}

	fmt.Println(resp.Status)
	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(resp.Header.Get("X-Custom-Header"))

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}
	fmt.Println(string(body))

	// Output:
	// 200 OK
	// text/plain
	// custom header from mock
	// hello, world
}

func ExampleHTTPMockerWithCustomHandler() {
	server := httpmocker.Launch(
		httpmocker.Response{
			Method: "GET",
			Path:   "/hello",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				io.WriteString(w, "hello, world from custom handler")
			},
		},
	)
	defer server.Close()

	url := fmt.Sprintf("%s/hello", server.URL)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}

	fmt.Println(resp.Status)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}
	fmt.Println(string(body))

	// Output:
	// 200 OK
	// hello, world from custom handler
}

func ExampleHTTPMockerWithUnknownHandler() {
	server := httpmocker.Launch()
	server.UnknownRequestHandler = func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "not found from unknown handler")
	}
	defer server.Close()

	url := fmt.Sprintf("%s/sushi", server.URL)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}

	fmt.Println(resp.Status)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}
	fmt.Println(string(body))

	// Output:
	// 404 Not Found
	// not found from unknown handler
}
