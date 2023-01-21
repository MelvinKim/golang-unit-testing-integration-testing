package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_application_handlers(t *testing.T) {
	var theTests = []struct {
		name               string
		url                string
		expectedStatusCode int
	}{
		{"home", "/", http.StatusOK},
		{"404", "/fish", http.StatusNotFound},
	}

	routes := app.routes()

	// creates a test server
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	// range through the tests
	for _, e := range theTests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for %s: expected status %d, but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestAppHome(t *testing.T) {
	var tests = []struct {
		name         string
		putInSession string
		expectedHTML string
	}{
		{"first visit", "", "<small>From session:"},
		{"second visit", "hello, world!", "<small>From session: hello, world!"},
	}

	for _, e := range tests {
		// create a request
		req, _ := http.NewRequest("GET", "/", nil)

		req = addContextAndSessionToRequest(req, app)
		_ = app.Session.Destroy(req.Context()) // make sure that the session is empty before running the tests

		if e.putInSession != "" {
			// add a value to the session
			app.Session.Put(req.Context(), "test", e.putInSession)
		}

		// we need a response
		rr := httptest.NewRecorder()

		// create the handler
		handler := http.HandlerFunc(app.Home)

		// serve the request
		handler.ServeHTTP(rr, req)

		// check status code
		if rr.Code != http.StatusOK {
			t.Errorf("TestAppHome returned wrong status code,expected %d, but got %d", http.StatusOK, rr.Code)
		}

		// check to ensure that the body contains a certain text that should be there
		body, _ := io.ReadAll(rr.Body)
		if !strings.Contains(string(body), e.expectedHTML) {
			t.Errorf("%s: did not find %s in response body", e.name, e.expectedHTML)
		}
	}
}

// Test a situation where i have a bad template
// func TestApp_renderWithBadTemplate(t *testing.T) {
// 	// set pathToTemplates to a location with a bad template
// 	pathToTemplates = "./testdata/"

// 	req, _ := http.NewRequest("GET", "/", nil)

// 	req = addContextAndSessionToRequest(req, app)

// 	rr := httptest.NewRecorder()

// 	err := app.render(rr, req, "bad.page.gohtml", &TemplateData{})

// 	if err == nil {
// 		t.Error("expected error from bad template, but did not get one")
// 	}

// 	pathToTemplates = "./../../templates"
// }

// adds "contextUserKey" value to the requests's context
// context.WithValue --> adds a new value to a context
func getCtx(req *http.Request) context.Context {
	ctx := context.WithValue(req.Context(), contextUserKey, "unknown")
	return ctx
}

// adds value to context and also adds sessions data to the context
func addContextAndSessionToRequest(req *http.Request, app application) *http.Request {
	req = req.WithContext(getCtx(req))

	ctx, _ := app.Session.Load(req.Context(), req.Header.Get("X-Session"))

	return req.WithContext(ctx)
}
