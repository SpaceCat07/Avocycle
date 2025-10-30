package utils

import "github.com/gin-gonic/gin"

// SuccessResponse bikin response sukses standar
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
    c.JSON(statusCode, gin.H{
        "success": true,
        "message": message,
        "data":    data,
    })
}

// ErrorResponse bikin response error standar
func ErrorResponse(c *gin.Context, statusCode int, message string, detail interface{}) {
    c.JSON(statusCode, gin.H{
        "success": false,
        "message": message,
        "error":   detail,
    })
}
