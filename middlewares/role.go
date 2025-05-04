package middlewares

import (
	"net/http"
	"strings"

	"abs-be/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gin-gonic/gin"
)

func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userClaims, exists := c.Get("user")
		if !exists {
			utils.ErrorResponse(c, http.StatusForbidden, "akses ditolak: user tidak ditemukan di context")
			c.Abort()
			return
		}

		claimsMap, ok := userClaims.(jwt.MapClaims)

		if !ok {
			utils.ErrorResponse(c, http.StatusForbidden, "akses ditolak: format claims tidak valid")
			c.Abort()
			return
		}

		userRole, ok := claimsMap["role"].(string)
		if !ok {
			utils.ErrorResponse(c, http.StatusForbidden, "akses ditolak: role tidak tersedia")
			c.Abort()
			return
		}

		allowedRolesStr := strings.Join(allowedRoles, ", ")

		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		utils.ErrorResponse(c, http.StatusForbidden, "akses ditolak: role Anda ('"+userRole+"') tidak diizinkan. Role yang diizinkan adalah: "+allowedRolesStr)
		c.Abort()
	}
}
