package infrastucture

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
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

		jwksJSON := retrievePublicKey(c, config.Auth.Domain)

		k, err := keyfunc.NewJWKSetJSON(jwksJSON)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Message": "Could not retrieve auth public key"})
		}

		token, err := jwt.ParseWithClaims(
			tokenString,
			&Claims{},
			k.Keyfunc,
			jwt.WithIssuer("https://"+config.Auth.Domain+"/"),
			jwt.WithAudience(config.Auth.Audience),
			jwt.WithValidMethods([]string{"RS256"}),
		)

		if err != nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
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

func retrievePublicKey(c *gin.Context, domain string) json.RawMessage {
	serverUrl := "https://" + domain + "/.well-known/jwks.json"
	resp, err := http.Get(serverUrl)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"Message": "Auth server connectivity malfucntion"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Message": "Auth server unexpected behavior"})
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Message": "Auth server connectivity malfucntion"})
	}

	return json.RawMessage(body)
}
