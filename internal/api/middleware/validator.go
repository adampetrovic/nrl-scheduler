package middleware

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func RequestValidator(validate *validator.Validate) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store validator in context for use in handlers
		c.Set("validator", validate)
		c.Next()
	}
}

// BindAndValidate binds request JSON and validates it
func BindAndValidate(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return err
	}
	
	validate, exists := c.Get("validator")
	if !exists {
		return nil // No validator configured
	}
	
	v := validate.(*validator.Validate)
	
	// Register custom tag name function to use JSON tags
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	
	return v.Struct(obj)
}

// BindQueryAndValidate binds query parameters and validates them
func BindQueryAndValidate(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return err
	}
	
	validate, exists := c.Get("validator")
	if !exists {
		return nil // No validator configured
	}
	
	v := validate.(*validator.Validate)
	return v.Struct(obj)
}