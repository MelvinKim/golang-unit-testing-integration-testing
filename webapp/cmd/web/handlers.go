package main

import (
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

func (app *application) Profile(w http.ResponseWriter, r *http.Request) {
	_ = app.render(w, r, "profile.page.gohtml", &TemplateData{})
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
		// put some data into the session
		app.Session.Put(r.Context(), "error", "invalid login credentials")
		// redirect to the login page, with an error message
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// get the email and password
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := app.DB.GetUserByEmail(email)
	if err != nil {
		app.Session.Put(r.Context(), "error", "invalid login!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	log.Println(password, user.FirstName)

	// after getting the user, we want to authenticate
	// if not authenticated then redirect with error

	// prevent fixation attack --> renew the session token
	_ = app.Session.RenewToken(r.Context())

	// store success message in the session

	// redirect to a profile page
	app.Session.Put(r.Context(), "flash", "Successfully logged in!")
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}
