package app

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"message": message,
	})
	c.Abort()
}

func SuccessResponse(c *gin.Context, content gin.H, message string) {
	c.JSON(http.StatusOK, content)
}
