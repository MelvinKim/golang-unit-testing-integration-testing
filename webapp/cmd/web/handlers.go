package main

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"time"
	"webapp/pkg/data"
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
	IP    string
	Data  map[string]any
	Error string
	Flash string
	User  data.User
}

func (app *application) render(w http.ResponseWriter, r *http.Request, t string, td *TemplateData) error {
	// always parse the template from disk
	parsedTemplate, err := template.ParseFiles(path.Join(pathToTemplates, t), path.Join(pathToTemplates, "base.layout.gohtml"))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
	}

	// get the client request IP
	td.IP = app.ipFromContext(r.Context())

	// get the error, if any, from the session --> after retrieving the value, it's deleted from the session
	td.Error = app.Session.PopString(r.Context(), "error")

	// get the flash message, if any, from the session --> after retrieving the value, it's deleted from the session
	td.Flash = app.Session.PopString(r.Context(), "flash")

	// execute the template, passing it data if any
	err = parsedTemplate.Execute(w, td)
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

	// after getting the user, we want to authenticate
	// if not authenticated then redirect with error
	if !app.authenticate(r, user, password) {
		app.Session.Put(r.Context(), "error", "invalid login!")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// prevent fixation attack (OWASP security guidelines) --> renew the session token
	_ = app.Session.RenewToken(r.Context())

	// store success message in the session
	// redirect to a profile page
	app.Session.Put(r.Context(), "flash", "Successfully logged in!")
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}

// authenticate compares provided password and the hashed user password
func (app *application) authenticate(r *http.Request, user *data.User, password string) bool {
	if valid, err := user.PasswordMatches(password); err != nil || !valid {
		return false
	}

	// store authenticated user in the session
	app.Session.Put(r.Context(), "user", user)

	return true
}
