package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dxbednarczyk/browsrr/internal/providers"
	"github.com/dxbednarczyk/browsrr/templates"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.POST("/query/", func(ctx echo.Context) error {
		err := ctx.Request().ParseForm()
		if err != nil {
			return ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to parse query data: %v", err))
		}

		provider := ctx.FormValue("provider")

		switch provider {
		case "1337x":
			return providers.One337X(ctx)
		case "nyaa":
			return providers.Nyaa(ctx, false)
		case "sukebei":
			return providers.Nyaa(ctx, true)
		default:
			return ctx.String(http.StatusInternalServerError, "invalid provider selected")
		}
	})

	e.POST("/error", func(ctx echo.Context) error {
		body, err := io.ReadAll(ctx.Request().Body)
		if err != nil {
			return err
		}

		errs := []error{errors.New(string(body))}

		var buf bytes.Buffer
		templates.Errors(errs).Render(context.Background(), &buf)

		return ctx.HTML(http.StatusOK, buf.String())
	})

	e.GET("/", func(ctx echo.Context) error {
		return templates.Render(ctx, http.StatusOK, templates.Index())
	})

	e.Logger.Fatal(e.Start(":3000"))
}
