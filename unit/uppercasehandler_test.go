package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpperCaseHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/upper?word=abc", nil)
	w := httptest.NewRecorder()
	upperCaseHandler(w, req)
	res := w.Result()

	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if string(data) != "ABC" {
		t.Errorf("expected ABC got %v", string(data))
	}
}

func TestClientUpperCase(t *testing.T) {
	expected := "Dummy data"
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, expected)
	}))

	defer svr.Close()

	c := NewClient(svr.URL)
	res, err := c.UpperCase("anything")
	if err != nil {
		t.Errorf("expected err to be nil got %v", err)
	}

	res = strings.TrimSpace(res)
	if res != expected {
		t.Errorf("expected res to be %s got %s", expected, res)
	}
}

// Testing Go server handlers is relatively easy, especially when you want to test just the handler logic.
// You just need to mock http.ResponseWriter and http.Request objects in your tests.
