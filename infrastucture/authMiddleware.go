package infrastucture

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	"github.com/golobby/container/v3"
	adapter "github.com/gwatts/gin-adapter"
	"github.com/steadfastie/gokube/infrastucture/services"
	"go.uber.org/zap"
)

const (
	ReadCounterScope   = "read:counter"
	CreateCounterScope = "create:counter"
	UpdateCounterScope = "update:counter"
)

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Scope string `json:"scope"`
}

// Validate does nothing for this example, but we need
// it to satisfy validator.CustomClaims interface.
func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

func AuthMiddleware(c *gin.Context) {
	container.Call(func(config *services.Config, logger *zap.Logger) {
		nextHandler, wrapper := adapter.New()
		next := EnsureValidToken(config, logger)(nextHandler)
		wrapper(next)
	})
}

// EnsureValidToken is a middleware that will check the validity of our JWT.
func EnsureValidToken(config *services.Config, logger *zap.Logger) func(next http.Handler) http.Handler {
	issuerURL, err := url.Parse("https://" + config.Auth.Domain + "/")
	if err != nil {
		logger.Fatal("Failed to parse the issuer url: %v", zap.Error(err))
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{config.Auth.Audience},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		logger.Fatal("Failed to set up the jwt validator")
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Info("Encountered error while validating JWT: %v", zap.Error(err))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Failed to validate JWT."}`))
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return func(next http.Handler) http.Handler {
		return middleware.CheckJWT(next)
	}
}

// HasScope checks whether our claims have a specific scope.
func (c CustomClaims) HasScope(expectedScope string) bool {
	result := strings.Split(c.Scope, " ")
	for i := range result {
		if result[i] == expectedScope {
			return true
		}
	}

	return false
}

func RequireScope(scope string) gin.HandlerFunc {
	return func(gc *gin.Context) {
		token := gc.Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
		claims := token.CustomClaims.(*CustomClaims)
		if !claims.HasScope(scope) {
			gc.AbortWithStatus(http.StatusForbidden)
			return
		}
		gc.Next()
	}
}
