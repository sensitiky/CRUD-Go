package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"server-go/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var JwtSecret = []byte(os.Getenv("JWT_SECRET"))

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("Authorization header missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			log.Println("Bearer token missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token missing"})
			c.Abort()
			return
		}

		claims := &jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method != jwt.SigningMethodHS512 {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return JwtSecret, nil
		})

		if err != nil || !token.Valid {
			log.Printf("Invalid Token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
			c.Abort()
			return
		}

		userId, okID := (*claims)["user_id"].(float64)
		name, okName := (*claims)["user_Name"].(string)
		lastName, okLastName := (*claims)["user_LastName"].(string)
		email, okEmail := (*claims)["user_Email"].(string)

		if !okID || !okName || !okLastName || !okEmail {
			log.Printf("Invalid token data: userId: %v, name: %v, lastName: %v, email: %v", userId, name, lastName, email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token data"})
			c.Abort()
			return
		}

		c.Set("user", &models.User{
			Id:       int(userId),
			Name:     name,
			LastName: lastName,
			Email:    email,
		})

		log.Printf("Token validated successfully: userId: %v, name: %v, lastName: %v, email: %v", userId, name, lastName, email)
		c.Next()
	}
}
