package main

import (
	"net/http"

	"github.com/dxbednarczyk/browsrr/internal/providers"
	"github.com/dxbednarczyk/browsrr/templates"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.SetTrustedProxies(nil)
	r.HTMLRender = &TemplRender{}

	r.POST("/query/", func(ctx *gin.Context) {
		err := ctx.Request.ParseForm()
		if err != nil {
			ctx.String(http.StatusInternalServerError, "failed to parse query data")

			ctx.AbortWithStatus(http.StatusInternalServerError)
		}

		provider := ctx.PostForm("provider")

		switch provider {
		case "1337x":
			providers.One337X(ctx)
		case "nyaa":
			providers.Nyaa(ctx)
		default:
			ctx.String(http.StatusInternalServerError, "invalid provider selected")

			ctx.AbortWithStatus(http.StatusInternalServerError)
		}
	})

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "", templates.Index())
	})

	r.Run(":3000")
}
