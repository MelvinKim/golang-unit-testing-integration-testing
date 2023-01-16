package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net/http"
	"webapp/pkg/data"
	"webapp/pkg/db"

	"github.com/alexedwards/scs/v2"
)

type application struct {
	DSN string
	// DB      *sql.DB
	DB      db.PostgresConn
	Session *scs.SessionManager
}

func main() {
	// register user defined types (eg Structs), so as to use store them in the session
	gob.Register(data.User{})

	// set up an app config --> variable that holds the application configuration
	// holds information that we want to share through out the applucation
	app := application{}

	// read something from the cmd
	flag.StringVar(&app.DSN, "dsn", "host=localhost port=5433 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5", "Postgres connection")
	flag.Parse()

	// try to connect to the database
	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	app.DB = db.PostgresConn{DB: conn}

	// get a session Manager --> should happen before invoking the routes
	app.Session = getSession()

	// get application routes
	mux := app.routes()

	// print out a message, (just to say the application is starting) :)
	log.Println("starting server on port 8080...")

	// start the server
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}

// advantages of 3rd routing packages
// faster that standard http library
// adds a lot of functionality with very little effort
