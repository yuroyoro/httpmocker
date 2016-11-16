httpmocker
================================
easy http response mocking by net/httptest

[![Build Status](https://travis-ci.org/yuroyoro/httpmocker.svg)](https://travis-ci.org/yuroyoro/httpmocker) [![GoDoc](https://godoc.org/github.com/yuroyoro/httpmocker?status.svg)](https://godoc.org/github.com/yuroyoro/httpmocker)

Examples:

https://github.com/yuroyoro/httpmocker/blob/master/example_test.go

### simple mocking

```go
	server := httpmocker.Launch().Add(
		"GET",
		"/hello",
		http.StatusOK,
		"hello, world",
	).Add(
		"POST",
		"/sushi",
		http.StatusCreated,
		"🍣",
	)
	defer server.Close()

	url := fmt.Sprintf("%s/hello", server.URL)
	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}

	fmt.Println(resp.Status)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}
	fmt.Println(string(body))
	resp.Body.Close()

	url = fmt.Sprintf("%s/sushi", server.URL)
	resp, err = http.Post(url, "text/plain", nil)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}

	fmt.Println(resp.Status)

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}
	fmt.Println(string(body))

	// Output:
	// 200 OK
	// hello, world
	// 201 Created
	// 🍣
```


### mocking with custom header

```
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
	defer resp.Body.Close()

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
```

### mocking with multple response with query strings

```
	server := httpmocker.Launch(
		httpmocker.Response{
			Method: "GET",
			Path:   "/hello",
			Code:   http.StatusOK,
			Body:   "hello, world",
		},
		httpmocker.Response{
			Method: "GET",
			Path:   "/hello",
			Query:  "dummy=1",
			Code:   http.StatusOK,
			Body:   "hello, world with query string",
		},
	)
	defer server.Close()

	// if no query string is given, mock server should return first mock response
	url := fmt.Sprintf("%s/hello", server.URL)
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}
	fmt.Println(string(body))
	resp.Body.Close()

	// if query string given and matched, mock server should return second mock response
	url = fmt.Sprintf("%s/hello?dummy=1", server.URL)
	resp, err = http.Get(url)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}
	fmt.Println(string(body))
	resp.Body.Close()

	// if query string given and not matched, mock server should return first mock response
	url = fmt.Sprintf("%s/hello?dummy=2", server.URL)
	resp, err = http.Get(url)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("unexpected error : %+v", err)
	}
	fmt.Println(string(body))

	// Output:
	// hello, world
	// hello, world with query string
	// hello, world
```

### mocking with custom response handler


```
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
	defer resp.Body.Close()
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
```

### mocking with unknown request handler

```
	server := httpmocker.Launch()
	server.UnknownRequestHandler = func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "not found from unknown handler")
	}
	defer server.Close()

	url := fmt.Sprintf("%s/sushi", server.URL)
	resp, err := http.Get(url)
	defer resp.Body.Close()
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
```
