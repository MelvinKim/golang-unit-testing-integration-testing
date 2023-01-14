package main

import (
	"log"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

type application struct {
	Session *scs.SessionManager
}

func main() {
	// set up an app config --> variable that holds the application configuration
	// holds information that we want to share through out the applucation
	app := application{}

	// get a session Manager --> should happen before invoking the routes
	app.Session = getSession()

	// get application routes
	mux := app.routes()

	// print out a message, (just to say the application is starting) :)
	log.Println("starting server on port 8080...")

	// start the server
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}

// advantages of 3rd routing packages
// faster that standard http library
// adds a lot of functionality with very little effort
