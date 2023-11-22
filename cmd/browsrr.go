package main

import (
	"net/http"

	"github.com/dxbednarczyk/browsrr/internal/providers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/1337x/{encoded}", providers.One337X)

	http.ListenAndServe(":3000", r)
}
