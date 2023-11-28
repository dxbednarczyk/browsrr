package providers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dxbednarczyk/browsrr/templates"
	"github.com/labstack/echo/v4"
)

func Sukebei(ctx echo.Context) error {
	query := ctx.FormValue("query")
	query = strings.Trim(query, " ")

	formatted := fmt.Sprintf("https://sukebei.nyaa.si/?q=%s", query)

	r := templates.NyaaResult{}

	doc, err := scrapeSite(formatted)
	if err != nil {
		r.Errors = append(r.Errors, fmt.Errorf("failed to scrape site: %v", err))

		return templates.Render(ctx, http.StatusInternalServerError, templates.NyaaResultTemplate(&r))
	}

	parseNyaaDocument(doc, &r)

	statusCode := http.StatusOK
	if r.Errors != nil {
		statusCode = http.StatusConflict
	}

	return templates.Render(ctx, statusCode, templates.NyaaResultTemplate(&r))
}
