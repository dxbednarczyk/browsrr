package main

import (
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/dxbednarczyk/browsrr/internal/providers"
	"github.com/dxbednarczyk/browsrr/templates"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(httprate.LimitByIP(3, 5*time.Second))

	r.Post("/query/", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse query data"))
		}

		provider := r.PostForm.Get("provider")

		switch provider {
		case "1337x":
			providers.One337X(w, r)
		case "nyaa":
			providers.Nyaa(w, r)
		default:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid provider selected"))
		}
	})

	index := templates.Index()
	r.Get("/", templ.Handler(index).ServeHTTP)

	http.ListenAndServe(":3000", r)
}
