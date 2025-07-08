//go:build prod

package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed all:web/dist
var webFS embed.FS

func serveUI(engine *gin.Engine) {
	dist, err := fs.Sub(webFS, "web/dist")
	if err != nil {
		panic(err)
	}
	engine.StaticFS("/assets", http.FS(mustFS(fs.Sub(dist, "assets"))))

	// Serve index.html for root and any other routes not handled by API
	engine.StaticFileFS("/", "index.html", http.FS(dist))
	engine.NoRoute(func(c *gin.Context) {
		c.FileFromFS("index.html", http.FS(dist))
	})
}

func mustFS(fs fs.FS, err error) fs.FS {
	if err != nil {
		panic(err)
	}
	return fs
}
