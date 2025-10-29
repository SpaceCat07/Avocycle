package middleware

import (
	"Avocycle/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// get the token from header
		authToken := ctx.GetHeader("Authorization")

		// check the "Bearer " string
		if authToken == "" || !strings.HasPrefix(authToken, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error" : "Missing or invalid authorization header"})
			return 
		}

		// trim the prefix
		tokenString := strings.TrimPrefix(authToken, "Bearer ")

		// get the data from token
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Check if user role is in allowed roles
        hasPermission := false
        for _, allowedRole := range allowedRoles {
            if claims.Role == allowedRole {
                hasPermission = true
                break
            }
        }

        if !hasPermission {
            ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
                "success": false,
                "error":   "Forbidden - insufficient role permission",
            })
            return
        }

		// continue to the next
		ctx.Next()
	}
}