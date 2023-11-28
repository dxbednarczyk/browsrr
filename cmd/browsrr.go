package main

import (
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
			ctx.String(http.StatusInternalServerError, "failed to parse query data")

			return err
		}

		provider := ctx.FormValue("provider")

		switch provider {
		case "1337x":
			return providers.One337X(ctx)
		case "nyaa":
			return providers.Nyaa(ctx)
		case "sukebei":
			return providers.Sukebei(ctx)
		default:
			ctx.String(http.StatusInternalServerError, "invalid provider selected")
		}

		return nil
	})

	e.GET("/", func(ctx echo.Context) error {
		return templates.Render(ctx, http.StatusOK, templates.Index())
	})

	e.Logger.Fatal(e.Start(":3000"))
}
