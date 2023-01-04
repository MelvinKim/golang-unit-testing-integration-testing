package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	// register middlware
	mux.Use(middleware.Recoverer)

	// register routes
	mux.Get("/", app.Home)

	// serve static assets
	return mux
}
