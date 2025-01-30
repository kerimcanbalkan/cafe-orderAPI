package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/kerimcanbalkan/cafe-orderAPI/config"
)

func Authenticate(allowedRoles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from the request header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Bearer token format
		if !strings.HasPrefix(tokenString, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		// Remove 'Bearer ' prefix from the token string
		tokenString = tokenString[7:]

		// Parse the token
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			func(token *jwt.Token) (interface{}, error) {
				return []byte(config.Env.Secret), nil
			},
		)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Get the role from the token
		role, ok := claims["Role"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Role not found in token"})
			c.Abort()
			return
		}

		// Check if the user's role is in the allowed roles
		roleIsValid := false
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				roleIsValid = true
				break
			}
		}

		if !roleIsValid {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		// Continue to the next middleware/handler if role matches
		c.Set("claims", claims)
		c.Next()
	}
}
