package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// Test to check whether the Form has a particular field
func TestForm_Has(t *testing.T) {
	form := NewForm(nil)

	has := form.Has("whatever")
	if has {
		t.Errorf("form shows it has field %s, when it should not", "whatever")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	form = NewForm(postedData)

	has = form.Has("a")
	if !has {
		t.Errorf("shows form doesn't have the field %s, when it should", "a")
	}
}

// Test to check for required fields
func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := NewForm(r.PostForm)

	// create required fields
	form.Required("a", "b", "c")
	if form.Valid() {
		t.Errorf("shows form valid when required fields %s, %s and %s are missing", "a", "b", "c")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	r, _ = http.NewRequest("POST", "/whatever", nil)
	r.PostForm = postedData

	form = NewForm(r.PostForm)
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Errorf("shows post doesn't have required fields %s, %s and %s when it does", "a", "b", "c")
	}
}

func TestForm_Check(t *testing.T) {
	form := NewForm(nil)

	form.Check(false, "password", "password is required")
	if form.Valid() {
		t.Errorf("Valid() returns %s, and it should be %s, when calling Check()", "false", "true")
	}
}

func TestForm_ErrorGet(t *testing.T) {
	form := NewForm(nil)
	form.Check(false, "password", "password is required")
	s := form.Errors.Get("password")

	if len(s) == 0 {
		t.Error("should have an error returned from Get, but does not")
	}

	s = form.Errors.Get("whatever")
	if len(s) != 0 {
		t.Error("should not have an error, but got one")
	}
}
