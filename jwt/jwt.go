package jwt

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ContextKey string

const (
	// UserIDContextKey is the default context key under which user ID string is stored.
	UserIDContextKey ContextKey = "user_id"
	// ScopeIDContextKey is the context key under which an optional scope/org/squad ID is stored.
	ScopeIDContextKey ContextKey = "scope_id"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrMissingToken = errors.New("missing authorization token")
)

// Claims represents the standard JWT claims model.
type Claims struct {
	UserID    string `json:"sub"`
	ScopeID   string `json:"scope_id,omitempty"`
	TokenType string `json:"type,omitempty"`
	jwt.RegisteredClaims
}

// Generate issues a signed HS256 JWT access token for the given user ID.
func Generate(userID string, secret []byte, ttl time.Duration, scopeID ...string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	if len(scopeID) > 0 && scopeID[0] != "" {
		claims.ScopeID = scopeID[0]
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// GenerateRefreshToken issues a signed HS256 JWT refresh token.
func GenerateRefreshToken(userID string, secret []byte, ttl time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(ttl)
	claims := Claims{
		UserID:    userID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	return signed, expiresAt, err
}

// Verify parses and verifies a JWT token string against the given HMAC secret.
func Verify(tokenStr string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrInvalidToken
}

// VerifyRefreshToken validates that the token is valid, unexpired, and has type "refresh".
func VerifyRefreshToken(tokenStr string, secret []byte) (string, error) {
	claims, err := Verify(tokenStr, secret)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}
	if claims.TokenType != "refresh" {
		return "", errors.New("not a refresh token")
	}
	if claims.UserID == "" {
		return "", errors.New("invalid refresh token: missing user id")
	}
	return claims.UserID, nil
}

// MiddlewareOption configures the Auth middleware.
type MiddlewareOption func(*middlewareConfig)

type middlewareConfig struct {
	scopeHeaders []string
	scopeParams  []string
	queryToken   bool
}

// WithScopeHeaders sets headers to inspect for scope/tenant/org/squad ID extraction.
func WithScopeHeaders(headers ...string) MiddlewareOption {
	return func(c *middlewareConfig) {
		c.scopeHeaders = headers
	}
}

// WithScopeQueryParams sets query parameter keys to inspect for scope/tenant/org/squad ID.
func WithScopeQueryParams(params ...string) MiddlewareOption {
	return func(c *middlewareConfig) {
		c.scopeParams = params
	}
}

// WithoutQueryToken disables extracting JWT tokens from ?token= URL query parameter.
func WithoutQueryToken() MiddlewareOption {
	return func(c *middlewareConfig) {
		c.queryToken = false
	}
}

// Middleware returns an HTTP authentication middleware validating HS256 JWT tokens.
func Middleware(secret []byte, opts ...MiddlewareOption) func(http.Handler) http.Handler {
	cfg := middlewareConfig{
		scopeHeaders: []string{"X-Organization-Id", "X-Squad-Id"},
		scopeParams:  []string{"org_id", "squad_id"},
		queryToken:   true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := ""

			// 1. Check Authorization: Bearer <token>
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenStr = parts[1]
				} else {
					tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			// 2. Fallback to ?token= query parameter (SSE EventSource / WebSockets)
			if tokenStr == "" && cfg.queryToken {
				tokenStr = r.URL.Query().Get("token")
			}

			if tokenStr == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := Verify(tokenStr, secret)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract scope / org ID
			scopeID := claims.ScopeID
			if scopeID == "" {
				for _, h := range cfg.scopeHeaders {
					if val := r.Header.Get(h); val != "" {
						scopeID = val
						break
					}
				}
			}
			if scopeID == "" {
				for _, p := range cfg.scopeParams {
					if val := r.URL.Query().Get(p); val != "" {
						scopeID = val
						break
					}
				}
			}

			ctx := context.WithValue(r.Context(), UserIDContextKey, claims.UserID)
			if scopeID != "" {
				ctx = context.WithValue(ctx, ScopeIDContextKey, scopeID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext retrieves the authenticated user ID string from context.
func UserIDFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(UserIDContextKey).(string); ok {
		return val
	}
	return ""
}

// ScopeIDFromContext retrieves the scope / org / squad ID string from context.
func ScopeIDFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(ScopeIDContextKey).(string); ok {
		return val
	}
	return ""
}
