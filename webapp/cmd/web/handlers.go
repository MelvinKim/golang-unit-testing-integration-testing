package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"time"
)

var pathToTemplates = "./templates/"

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	var td = make(map[string]any) //td --> templateData

	if app.Session.Exists(r.Context(), "test") {
		// retrieve data from the session
		msg := app.Session.GetString(r.Context(), "test")
		td["test"] = msg
	} else {
		// add something to the session
		app.Session.Put(r.Context(), "test", "Hit this page at "+time.Now().UTC().String())
	}
	_ = app.render(w, r, "home.page.gohtml", &TemplateData{Data: td})
}

type TemplateData struct {
	IP   string
	Data map[string]any
}

func (app *application) render(w http.ResponseWriter, r *http.Request, t string, data *TemplateData) error {
	// always parse the template from disk
	parsedTemplate, err := template.ParseFiles(path.Join(pathToTemplates, t), path.Join(pathToTemplates, "base.layout.gohtml"))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
	}

	// get the client request IP
	data.IP = app.ipFromContext(r.Context())

	// execute the template, passing it data if any
	err = parsedTemplate.Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	// for every POST action, parse the Form data
	err := r.ParseForm()
	if err != nil {
		log.Printf("error while parsing form data: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// validate data
	form := NewForm(r.PostForm)
	form.Required("email", "password")

	if !form.Valid() {
		fmt.Fprint(w, "failed input validation")
		return
	}

	// get the email and password
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := app.DB.GetUserByEmail(email)
	if err != nil {
		log.Println(err)
	}

	log.Println("From database: ", user.FirstName)

	log.Println(email, password)

	fmt.Fprint(w, email)
}
