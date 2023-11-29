package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/dxbednarczyk/browsrr/internal/providers"
	"github.com/dxbednarczyk/browsrr/templates"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
)

func main() {
	r := chi.NewRouter()

	styles := log.DefaultStyles()
	styles.Message = lipgloss.NewStyle().
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("37"))

	handler := log.New(os.Stdout)
	handler.SetStyles(styles)

	logger := slog.New(handler)

	httpLogger := httplog.Logger{
		Logger: logger,
		Options: httplog.Options{
			LogLevel: slog.LevelInfo,
			Concise:  true,
		},
	}

	r.Use(httplog.RequestLogger(&httpLogger))
	r.Use(middleware.Recoverer)

	r.Post("/query/", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("failed to parse query data: %v", err)))

			return
		}

		provider := r.FormValue("provider")

		switch provider {
		case "1337x":
			providers.One337X(w, r)
		case "nyaa":
			providers.Nyaa(w, r, false)
		case "sukebei":
			providers.Nyaa(w, r, true)
		default:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid provider selected"))
		}
	})

	r.Post("/error", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("failed to parse error data: %v", err)))

			return
		}

		errs := []error{errors.New(string(body))}

		h := templ.Handler(templates.Errors(errs))
		h.ServeHTTP(w, r)
	})

	r.Get("/", templ.Handler(templates.Index()).ServeHTTP)

	err := http.ListenAndServe(":3000", r)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
