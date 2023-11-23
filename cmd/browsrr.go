package main

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/dxbednarczyk/browsrr/internal/providers"
	"github.com/dxbednarczyk/browsrr/templates"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/1337x/", providers.One337X)

	index := templates.Index()
	r.Get("/", templ.Handler(index).ServeHTTP)

	http.ListenAndServe(":3000", r)
}
