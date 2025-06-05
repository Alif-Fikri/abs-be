package middlewares

import (
	"net/http"
	"os"
	"strings"

	"abs-be/database"
	"abs-be/models"
	"abs-be/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "authorization header tidak ditemukan")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "authorization header tidak valid")
			c.Abort()
			return
		}

		tokenString := parts[1]

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			utils.ErrorResponse(c, http.StatusInternalServerError, "jwt secret tidak tersedia")
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			utils.ErrorResponse(c, http.StatusUnauthorized, "token tidak valid")
			c.Abort()
			return
		}

		var session models.Session
		if err := database.DB.Where("token = ?", tokenString).First(&session).Error; err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "sesi login tidak ditemukan atau sudah logout")
			c.Abort()
			return
		}

		// if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// 	c.Set("user", claims)
		// 	c.Set("tokenString", tokenString)
		// }
		c.Set("user_id", session.UserID)
		c.Set("role", session.Role)
		c.Next()
	}
}
