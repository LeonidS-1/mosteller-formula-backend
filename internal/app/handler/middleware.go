package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const bearerPrefix = "Bearer"

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (h *Handler) AuthMiddleware(requireModerator bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractTokenFromHeader(c.Request)
		if tokenString == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		jwtKey := os.Getenv("JWT_KEY")
		if jwtKey == "" {
			jwtKey = "default-secret-key-change-in-production"
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtKey), nil
		})
		if err != nil || token == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		blacklisted, err := h.Repository.IsTokenBlacklisted(context.Background(), tokenString)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if blacklisted {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userIDStr, ok := claims["user_id"].(string)
		if !ok || userIDStr == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		isModerator, ok := claims["is_moderator"].(bool)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if requireModerator && !isModerator {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Set("user_id", userIDStr)
		c.Set("is_moderator", isModerator)
		c.Next()
	}
}

func (h *Handler) WithOptionalAuthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractTokenFromHeader(c.Request)
		if tokenString == "" {
			c.Set("user_id", "")
			c.Next()
			return
		}

		blacklisted, err := h.Repository.IsTokenBlacklisted(context.Background(), tokenString)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if blacklisted {
			c.Set("user_id", "")
			c.Next()
			return
		}

		jwtKey := os.Getenv("JWT_KEY")
		if jwtKey == "" {
			jwtKey = "default-secret-key-change-in-production"
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtKey), nil
		})
		if err != nil || token == nil || !token.Valid {
			c.Set("user_id", "")
			c.Next()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Set("user_id", "")
			c.Next()
			return
		}

		userIDValue, exists := claims["user_id"]
		if !exists || userIDValue == nil {
			c.Set("user_id", "")
			c.Next()
			return
		}

		switch v := userIDValue.(type) {
		case string:
			c.Set("user_id", v)
		case float64:
			c.Set("user_id", fmt.Sprintf("%.0f", v))
		default:
			c.Set("user_id", fmt.Sprintf("%v", v))
		}

		c.Next()
	}
}

func extractTokenFromHeader(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if bearerToken == "" {
		return ""
	}

	parts := strings.SplitN(bearerToken, " ", 2)
	if len(parts) != 2 || parts[0] != bearerPrefix {
		return ""
	}
	return parts[1]
}

func getUserID(ctx *gin.Context) (uint, error) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists || userIDVal == nil {
		return 0, fmt.Errorf("user_id not found")
	}

	userIDStr, ok := userIDVal.(string)
	if !ok || userIDStr == "" {
		return 0, fmt.Errorf("user_id not found")
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(userID), nil
}
