# quikstop

`quikstop` is a cohesive, lightweight application development toolkit for Go. It provides standard, zero-overhead building blocks for web backends without heavy runtime frameworks or reflection-based DI.

## Quick Start (Single Import)

Every building block is re-exported at the root of `quikstop`. You only need to import one package in your application:

```go
package main

import (
	"context"
	"net/http"
	"time"

	"github.com/v-e-r-n/quikstop"
)

func main() {
	// 1. Structured Logging (all subpackages route through this)
	quikstop.SetLogger(quikstop.Logger())

	// 2. Configuration Helpers
	jwtSecret := []byte(quikstop.GetEnv("JWT_SECRET", "super-secret-key-1234567890123456"))
	addr := quikstop.ResolveAddr("HOST", []string{"PORT", "HTTP_PORT"}, "8080")

	// 3. Router & Middleware
	mux := http.NewServeMux()
	limiter := quikstop.NewLimiter(10.0, 20.0) // 10 req/s, burst 20
	auth := quikstop.AuthMiddleware(jwtSecret)

	mux.HandleFunc("POST /api/v1/login", func(w http.ResponseWriter, r *http.Request) {
		token, _ := quikstop.GenerateJWT("user-123", jwtSecret, 24*time.Hour)
		quikstop.JSON(w, http.StatusOK, map[string]string{"token": token})
	})

	mux.Handle("GET /api/v1/profile", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := quikstop.UserIDFromContext(r.Context())
		quikstop.JSON(w, http.StatusOK, map[string]string{"user_id": userID})
	})))

	// 4. Graceful Shutdown Server
	stack := quikstop.CORS()(quikstop.RateLimitMiddleware(limiter)(mux))
	srv := &http.Server{
		Addr:    addr,
		Handler: stack,
	}

	quikstop.ListenAndServeGracefully(srv, 5*time.Second)
}
```

---

## Modules & Aliases

```
github.com/v-e-r-n/quikstop
  ├── (root)      # Graceful HTTP server, env helpers & root re-exported aliases
  ├── core        # Global structured logger configuration used across all subpackages
  ├── cors        # Configurable CORS middleware with standard API defaults
  ├── db          # SQLite / PostgreSQL dialect detection, WAL tuning, query rebinding & Goose migrations
  ├── events      # Real-time Server-Sent Events (SSE) pub/sub, client tracking & targeted dispatching
  ├── httputil    # Standard JSON response, error payload, and body decoding helpers
  ├── jwt         # HS256 JWT generation, validation, and authentication middleware
  ├── limiter     # In-memory token-bucket IP rate limiting middleware with proxy IP parsing
  ├── mcfeely     # Transactional email delivery interface with Console and SMTP adapters
  └── otp         # Passwordless one-time passcode (OTP) generation, verification & HTTP handler
```

---

## 1. Core & Server

Environment variable helpers, shared structured logging, and graceful HTTP server lifecycle handling.

```go
// 1. Environment Variable Helpers
jwtSecret, err := quikstop.ReqEnv("JWT_SECRET") // errors if missing/empty
dbPath := quikstop.GetEnv("DATABASE_PATH", "data/app.db") // fallback default
addr := quikstop.ResolveAddr("HOST", []string{"PORT", "HTTP_PORT"}, "8080")

// 2. Graceful Shutdown Server
quikstop.ListenAndServeGracefully(srv, 5*time.Second)
```

---

## 2. Database (`quikstop.ConnectDB`)

Auto-detects database dialect (SQLite vs PostgreSQL) from connection string scheme, applies SQLite performance tuning (WAL, foreign keys), rebinds `?` placeholders to PostgreSQL `$1, $2` parameters dynamically, and embeds Goose migrations.

```go
//go:embed migrations/*.sql
var migrationFS embed.FS

database, err := quikstop.ConnectDB("postgres://user:pass@localhost:5432/app")
// Or: quikstop.ConnectDB("sqlite://data/app.db") or quikstop.ConnectDB("data/app.db")

// Run embedded goose migrations
err = database.Migrate(migrationFS, "migrations")

// Query rebinding: write SQL with '?' for both SQLite and Postgres
query := database.Rebind("SELECT id FROM users WHERE email = ? AND org_id = ?")
row := database.QueryRowContext(ctx, query, email, orgID)
```

---

## 3. JWT Authentication (`quikstop.AuthMiddleware`)

Generates, verifies, and extracts user/scope IDs from HS256 JWT tokens. The middleware automatically supports both standard `Authorization: Bearer <token>` headers and `?token=` query parameters (ideal for EventSource SSE / WebSocket transports).

```go
secret := []byte("your-secret-key")

// Issue a token
tokenStr, err := quikstop.GenerateJWT(userID, secret, 24*time.Hour, "optional-org-id")

// Protect routes
authMiddleware := quikstop.AuthMiddleware(secret,
	quikstop.WithScopeHeaders("X-Organization-Id", "X-Squad-Id"),
)

mux.Handle("GET /api/v1/profile", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	userID := quikstop.UserIDFromContext(r.Context())
	orgID := quikstop.ScopeIDFromContext(r.Context())
})))
```

---

## 4. Real-time Events & SSE (`quikstop.NewEventBus`, `quikstop.NewSSEHandler`)

Lightweight pub/sub bus and SSE connection dispatcher with targeted recipient resolution.

```go
bus := quikstop.NewEventBus()
sm := quikstop.NewStreamManager()

type MyResolver struct{}
func (r *MyResolver) ResolveClientIDs(ctx context.Context, ev quikstop.Event) ([]string, error) {
	return []string{"user-123"}, nil
}

dispatcher := quikstop.NewDispatcher(bus, sm, &MyResolver{})
dispatcher.Start(context.Background())

// Expose SSE endpoint
mux.Handle("GET /api/v1/events", quikstop.NewSSEHandler(sm, quikstop.SSEConfig{
	HeartbeatInterval: 25 * time.Second,
	ContextExtractor: func(r *http.Request) (quikstop.ClientSession, error) {
		return quikstop.ClientSession{
			ClientID: quikstop.UserIDFromContext(r.Context()),
			ScopeID:  quikstop.ScopeIDFromContext(r.Context()),
		}, nil
	},
}))

// Publish event anywhere in your app
bus.Publish(quikstop.NewEvent(
	quikstop.EventType("TASK_UPDATED"),
	quikstop.EventPayload{ResourceID: "task-456"},
	quikstop.EventMeta{ScopeID: "org-1"},
))
```

---

## 5. Rate Limiting (`quikstop.NewLimiter`)

Token-bucket IP rate limiter middleware with proxy-aware IP resolution (`CF-Connecting-IP`, `True-Client-IP`, `X-Forwarded-For`, `X-Real-IP`, and socket fallback) and periodic inactive IP cleanup.

```go
// 10 requests/second with a burst of 20
rateLimiter := quikstop.NewLimiter(10.0, 20.0, quikstop.WithOnLimit(func(ip string, r *http.Request) {
	log.Printf("Rate limit exceeded for IP %s on %s", ip, r.URL.Path)
}))

handler := quikstop.RateLimitMiddleware(rateLimiter)(mux)
```

---

## 6. Passwordless OTP (`quikstop.NewOTP`)

Manages OTP request, secure numeric code generation, expiration TTL, attempt count lockouts, and HTTP endpoints.

```go
store := quikstop.NewMemoryOTPStore()
deliverer := quikstop.NewConsoleDeliverer()
keeper := quikstop.NewOTP(store, deliverer, quikstop.OTPConfig{
	CodeLength: 6,
	TTL:        15 * time.Minute,
	MaxRetries: 3,
})

handler := quikstop.NewOTPHandler(keeper, func(w http.ResponseWriter, r *http.Request, recipient string) {
	// Called on successful verification: issue session / JWT
})

mux.HandleFunc("POST /auth/otp/request", handler.Knock)
mux.HandleFunc("POST /auth/otp/verify", handler.Verify)
```

---

## 7. Email Delivery (`quikstop.NewConsoleMcFeely`, `quikstop.NewSmtpMcFeely`)

Speedy delivery interface with standard Console and SMTP adapters.

```go
var mailer quikstop.McFeely

if isDev {
	mailer = quikstop.NewConsoleMcFeely()
} else {
	mailer = quikstop.NewSmtpMcFeely("smtp.example.com", "587", "user", "pass", "no-reply@example.com")
}

_ = mailer.Send(ctx, "user@example.com", "Welcome!", "Your account is ready.")
```

---

## 8. CORS (`quikstop.CORS`)

Preflight `OPTIONS` and cross-origin header handler.

```go
corsMiddleware := quikstop.CORS(
	quikstop.WithCORSOrigins("https://app.example.com"),
	quikstop.WithCORSMethods("GET", "POST", "PUT", "DELETE"),
)

http.ListenAndServe(":8080", corsMiddleware(mux))
```

---

## 9. HTTP Utilities (`quikstop.JSON`, `quikstop.Error`, `quikstop.DecodeJSON`)

Standard JSON response writers and error formatters.

```go
func HandleUser(w http.ResponseWriter, r *http.Request) {
	var body CreateUserRequest
	if err := quikstop.DecodeJSON(r, &body); err != nil {
		quikstop.Error(w, http.StatusBadRequest, "invalid_body", "Failed to parse JSON")
		return
	}

	user, err := CreateUser(body)
	if err != nil {
		quikstop.Error(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	quikstop.JSON(w, http.StatusCreated, user)
}
```

---

## Running Tests

```bash
go test -v ./...
```
