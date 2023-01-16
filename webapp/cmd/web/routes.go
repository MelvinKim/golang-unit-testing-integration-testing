package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	// register middlware, --> middlewares should go before the registered routes
	mux.Use(middleware.Recoverer)
	mux.Use(app.addIpToContext)
	mux.Use(app.Session.LoadAndSave) // add a middleware to persist sessions between requests

	// register routes
	mux.Get("/", app.Home)
	mux.Post("/login", app.Login)

	mux.Get("/user/profile", app.Profile)

	// serve static assets
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
