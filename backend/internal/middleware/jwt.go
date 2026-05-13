package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"backend/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const accessRefreshThreshold = 5 * time.Minute

func JWTAuthMiddleware(tokenService *auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tokenService == nil || !tokenService.HasSecret() {
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

			return []byte(tokenService.Secret()), nil
		})
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				if token == nil {
					logJWTError(err)
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
					return
				}

				claims, ok := token.Claims.(jwt.MapClaims)
				if !ok {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
					return
				}

				if deny, inGrace := tokenService.CheckAccessTokenLogout(getStringClaim(claims, "jti")); deny || inGrace {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
					return
				}

				if refreshed := refreshAccessToken(c, tokenService); !refreshed {
					logJWTError(err)
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
					return
				}
				c.Next()
				return
			}

			logJWTError(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}
		if !token.Valid {
			logJWTError(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		log.Printf("User Claims: ID=%v, Role=%v", claims["user_id"], claims["role"])

		jti := getStringClaim(claims, "jti")
		deny, inGrace := tokenService.CheckAccessTokenLogout(jti)
		if deny {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		setAuthContextFromClaims(c, claims, jti)
		c.Set("jwt", token)

		if !inGrace && shouldRefreshAccessToken(claims) {
			_ = refreshAccessToken(c, tokenService)
		}

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

func refreshAccessToken(c *gin.Context, tokenService *auth.TokenService) bool {
	refreshToken, err := c.Cookie(auth.RefreshTokenCookieName)
	if err != nil || refreshToken == "" {
		return false
	}

	access, identity, err := tokenService.RefreshAccessToken(refreshToken)
	if err != nil {
		return false
	}

	setAccessTokenHeader(c, access.Token)
	setAuthContext(c, identity, access.JTI)
	return true
}

func setAccessTokenHeader(c *gin.Context, token string) {
	if token == "" {
		return
	}

	c.Header(auth.AccessTokenHeaderName, token)
	c.Header("Access-Control-Expose-Headers", auth.AccessTokenHeaderName)
}

func setAuthContext(c *gin.Context, identity auth.TokenIdentity, jti string) {
	c.Set("user_id", identity.UserID)
	c.Set("role", identity.Role)
	c.Set("role_id", identity.RoleID)
	c.Set("jti", jti)
}

func setAuthContextFromClaims(c *gin.Context, claims jwt.MapClaims, jti string) {
	if userID, ok := getUintClaim(claims, "user_id"); ok {
		c.Set("user_id", userID)
	}

	if role, ok := claims["role"].(string); ok {
		c.Set("role", role)
	}

	if roleID, ok := getUintClaim(claims, "role_id"); ok {
		c.Set("role_id", roleID)
	}

	c.Set("jti", jti)
}

func shouldRefreshAccessToken(claims jwt.MapClaims) bool {
	expValue, ok := claims["exp"]
	if !ok {
		return false
	}

	expTime, ok := parseUnixTimestamp(expValue)
	if !ok {
		return false
	}

	return time.Until(expTime) <= accessRefreshThreshold
}

func parseUnixTimestamp(value interface{}) (time.Time, bool) {
	switch typed := value.(type) {
	case float64:
		return time.Unix(int64(typed), 0), true
	case float32:
		return time.Unix(int64(typed), 0), true
	case int:
		return time.Unix(int64(typed), 0), true
	case int64:
		return time.Unix(typed, 0), true
	case uint:
		return time.Unix(int64(typed), 0), true
	case uint64:
		return time.Unix(int64(typed), 0), true
	default:
		return time.Time{}, false
	}
}

func getStringClaim(claims jwt.MapClaims, key string) string {
	value, ok := claims[key]
	if !ok {
		return ""
	}

	if typed, ok := value.(string); ok {
		return typed
	}

	return ""
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
