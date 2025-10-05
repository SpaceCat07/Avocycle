package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware → batasi akses hanya untuk role tertentu
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role not found in token"})
			c.Abort()
			return
		}

		role := strings.ToLower(roleValue.(string)) // ✅ ubah jadi huruf kecil

		for _, allowed := range allowedRoles {
			if role == strings.ToLower(allowed) { // ✅ bandingkan lowercase
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		c.Abort()
	}
}
