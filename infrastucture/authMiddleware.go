package infrastucture

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golobby/container/v3"
	"github.com/steadfastie/gokube/infrastucture/services"
)

const (
	ReadCounterScope   = "read:counter"
	CreateCounterScope = "create:counter"
	UpdateCounterScope = "update:counter"
)

type Claims struct {
	Scope string `json:"scope"`
	jwt.RegisteredClaims
}

func AuthMiddleware(requiredScopes ...string) gin.HandlerFunc {
	var config *services.Config
	container.Resolve(&config)

	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader("Authorization")
		if authorizationHeader == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authorizationHeader, "Bearer ")

		token, err := jwt.ParseWithClaims(
			tokenString,
			&Claims{},
			func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "Unexpected token signing method"})
				}

				return []byte(config.Auth.Secret), nil
			},
			jwt.WithIssuer(config.Auth.Domain),
			jwt.WithAudience(config.Auth.Audience),
			jwt.WithValidMethods([]string{"HS256"}),
		)

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Can't recognize user"})
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims.HasScopes(requiredScopes...) {
			c.Set("user", claims.Subject)
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusForbidden)
		}
	}
}

func (c Claims) HasScopes(expectedScopes ...string) bool {
	result := strings.Fields(c.Scope)
	for _, expectedScope := range expectedScopes {
		for _, scope := range result {
			if scope == expectedScope {
				return true
			}
		}
	}
	return false
}
