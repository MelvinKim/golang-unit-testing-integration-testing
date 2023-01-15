package main

import (
	"os"
	"testing"
)

// hence this variable can be reused in the other test files
var app application

// will always be executed before the actual tests run
func TestMain(m *testing.M) {
	pathToTemplates = "./../../templates"

	// setup a session for our tests
	app.Session = getSession()

	os.Exit(m.Run())
}
