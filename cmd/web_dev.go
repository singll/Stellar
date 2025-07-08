//go:build !prod

package main

import "github.com/gin-gonic/gin"

// In development, Vite handles serving the UI.
// This is a no-op.
func serveUI(engine *gin.Engine) {
	// In development, the UI is served by the Vite dev server,
	// so the Go backend doesn't need to do anything.
}
