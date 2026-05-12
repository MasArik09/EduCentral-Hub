package middleware

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "JWT secret is not configured"})
			return
		}

		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		tokenStr := authHeader
		for i := 0; i < 2; i++ {
			if strings.HasPrefix(strings.ToLower(tokenStr), "bearer ") {
				tokenStr = strings.TrimSpace(tokenStr[len("bearer "):])
				continue
			}
			break
		}
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Bearer token is required"})
			return
		}

		parser := jwt.NewParser(jwt.WithLeeway(2 * time.Minute))
		token, err := parser.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}

			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			logJWTError(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		if userID, ok := getUintClaim(claims, "user_id"); ok {
			c.Set("user_id", userID)
		}

		if role, ok := claims["role"].(string); ok {
			c.Set("role", role)
		}

		if roleID, ok := getUintClaim(claims, "role_id"); ok {
			c.Set("role_id", roleID)
		}

		c.Set("jwt", token)
		c.Next()
	}
}

func logJWTError(err error) {
	if err == nil {
		log.Printf("jwt validation failed: token invalid")
		return
	}

	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		log.Printf("jwt validation failed: token expired (%v)", err)
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		log.Printf("jwt validation failed: signature mismatch (%v)", err)
	case errors.Is(err, jwt.ErrTokenMalformed):
		log.Printf("jwt validation failed: malformed token (%v)", err)
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		log.Printf("jwt validation failed: token not valid yet (%v)", err)
	default:
		log.Printf("jwt validation failed: %v", err)
	}
}

func getUintClaim(claims jwt.MapClaims, key string) (uint, bool) {
	value, ok := claims[key]
	if !ok {
		return 0, false
	}

	switch typed := value.(type) {
	case float64:
		return uint(typed), true
	case float32:
		return uint(typed), true
	case int:
		return uint(typed), true
	case int64:
		return uint(typed), true
	case uint:
		return typed, true
	case uint64:
		return uint(typed), true
	default:
		return 0, false
	}
}
