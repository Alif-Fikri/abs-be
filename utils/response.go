package utils

import (
	"github.com/gin-gonic/gin"
)

func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"success": false,
		"error":   message,
	})
}

func SuccessResponse(c *gin.Context, code int, message string, data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	c.JSON(code, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}
