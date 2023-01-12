package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_application_addIPToContext(t *testing.T) {
	tests := []struct {
		headerName  string
		headerValue string
		addr        string
		emptyAddr   bool
	}{
		{"", "", "", false},
		{"", "", "", true},
		{"X-Forwarded-For", "192.3.2.1", "", false},
		{"", "", "hello:world", false},
	}

	var app application

	// create a dummy hanlder that we will use to check the context
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// make sure that the value exists in the context
		val := r.Context().Value(contextUserKey)
		if val == nil {
			t.Error(contextUserKey, "not present")
		}

		//make sure we got a string back
		ip, ok := val.(string)
		if !ok {
			t.Error(val, "not a string")
		}
		t.Log(ip)
	})

	for _, e := range tests {
		// create the handler to test
		handlerToTest := app.addIpToContext(nextHandler)

		// create a request, in order to be able to perform our tests --> creates a Mock request
		req := httptest.NewRequest("GET", "http://testing", nil)

		// check for an empty address
		if e.emptyAddr {
			req.RemoteAddr = ""
		}

		if len(e.headerName) > 0 {
			req.Header.Add(e.headerName, e.headerValue)
		}

		if len(e.addr) > 0 {
			req.RemoteAddr = e.addr
		}

		handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
	}
}

func Test_application_ipFromContext(t *testing.T) {
	// create an app variable of type application
	var app application

	// get context
	ctx := context.Background()

	// put something in the context
	ctx = context.WithValue(ctx, contextUserKey, "192.0.0.1")

	// call the ipFromContext function
	ip := app.ipFromContext(ctx)

	// perform the test
	if !strings.EqualFold("192.0.0.1", ip) {
		t.Errorf("wrong value returned from context, got %s, want %s", ip, "192.0.0.1")
	}
}
