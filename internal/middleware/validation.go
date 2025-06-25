package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateJSON middleware validates JSON request body against a struct
func ValidateJSON(structType interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(structType); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_failed",
				"message": "Request body validation failed",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		if err := validate.Struct(structType); err != nil {
			validationErrors := make([]string, 0)
			for _, err := range err.(validator.ValidationErrors) {
				validationErrors = append(validationErrors, err.Error())
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_failed",
				"message": "Request body validation failed",
				"details": validationErrors,
			})
			c.Abort()
			return
		}
		c.Set("validated_body", structType)
		c.Next()
	}
}
