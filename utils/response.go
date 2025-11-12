package utils

import "github.com/gin-gonic/gin"

// Response structure untuk API response
type Response struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Error   any    `json:"error,omitempty"`
    Data    any    `json:"data,omitempty"`
    Meta    any    `json:"meta,omitempty"`
}

// EmptyObj untuk response tanpa data
type EmptyObj struct{}

// BuildResponse creates a custom response
func BuildResponse(success bool, message string, data any, meta any, err any) Response {
    return Response{
        Success: success,
        Message: message,
        Data:    data,
        Meta:    meta,
        Error:   err,
    }
}

// SuccessResponse sends success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
    c.JSON(statusCode, Response{
        Success: true,
        Message: message,
        Data:    data,
    })
}

// SuccessResponseWithMeta sends success response with metadata (for pagination)
func SuccessResponseWithMeta(c *gin.Context, statusCode int, message string, data interface{}, meta interface{}) {
    c.JSON(statusCode, Response{
        Success: true,
        Message: message,
        Data:    data,
        Meta:    meta,
    })
}

// ErrorResponse sends error response
func ErrorResponse(c *gin.Context, statusCode int, message string, detail interface{}) {
    c.JSON(statusCode, Response{
        Success: false,
        Message: message,
        Error:   detail,
    })
}

// ErrorResponseWithData sends error response with data
func ErrorResponseWithData(c *gin.Context, statusCode int, message string, data interface{}, detail interface{}) {
    c.JSON(statusCode, Response{
        Success: false,
        Message: message,
        Data:    data,
        Error:   detail,
    })
}

// BuildSuccessResponse creates success response object (without sending)
func BuildSuccessResponse(message string, data any) Response {
    return Response{
        Success: true,
        Message: message,
        Data:    data,
    }
}

// BuildSuccessResponseWithMeta creates success response with meta (without sending)
func BuildSuccessResponseWithMeta(message string, data any, meta any) Response {
    return Response{
        Success: true,
        Message: message,
        Data:    data,
        Meta:    meta,
    }
}

// BuildErrorResponse creates error response object (without sending)
func BuildErrorResponse(message string, err any) Response {
    return Response{
        Success: false,
        Message: message,
        Error:   err,
    }
}