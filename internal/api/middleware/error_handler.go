package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/adampetrovic/nrl-scheduler/pkg/types"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			switch e := err.Err.(type) {
			case validator.ValidationErrors:
				handleValidationError(c, e)
				return
			default:
				handleGenericError(c, e)
				return
			}
		}
	}
}

func handleValidationError(c *gin.Context, err validator.ValidationErrors) {
	details := make(map[string]string)
	
	for _, e := range err {
		field := e.Field()
		switch e.Tag() {
		case "required":
			details[field] = "This field is required"
		case "min":
			details[field] = "Value is too small"
		case "max":  
			details[field] = "Value is too large"
		case "email":
			details[field] = "Invalid email format"
		case "oneof":
			details[field] = "Invalid value"
		default:
			details[field] = "Invalid value"
		}
	}
	
	c.JSON(http.StatusBadRequest, types.ErrorResponse{
		Error:   "Validation failed",
		Code:    "VALIDATION_ERROR",
		Details: details,
	})
}

func handleGenericError(c *gin.Context, err error) {
	// Check if we already have a status code set
	if c.Writer.Status() != http.StatusOK {
		c.JSON(c.Writer.Status(), types.ErrorResponse{
			Error: err.Error(),
			Code:  "REQUEST_ERROR",
		})
		return
	}
	
	// Default to internal server error
	c.JSON(http.StatusInternalServerError, types.ErrorResponse{
		Error: "Internal server error",
		Code:  "INTERNAL_ERROR",
	})
}

// Helper functions for handlers to return errors easily
func BadRequest(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, types.ErrorResponse{
		Error: message,
		Code:  "BAD_REQUEST",
	})
}

func NotFound(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusNotFound, types.ErrorResponse{
		Error: message,
		Code:  "NOT_FOUND",
	})
}

func InternalError(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, types.ErrorResponse{
		Error: message,
		Code:  "INTERNAL_ERROR",
	})
}

func Conflict(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusConflict, types.ErrorResponse{
		Error: message,
		Code:  "CONFLICT",
	})
}