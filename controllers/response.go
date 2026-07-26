// Package controllers contains the HTTP handlers (the web layer). Handlers
// translate requests into service calls and marshal responses; they hold no
// business logic themselves.
package controllers

import "github.com/gin-gonic/gin"

// respondError writes a standard JSON error envelope with the given status.
func respondError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}
