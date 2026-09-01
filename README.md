# quikstop

`quikstop` is a cohesive, lightweight application development toolkit for Go. It provides standard, zero-overhead building blocks for web backends without heavy runtime frameworks or reflection-based DI.

## Modules & Packages

```
github.com/v-e-r-n/quikstop
  ├── (root)      # Graceful HTTP server shutdown and environment config helpers
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

## 1. Core (`quikstop`)

Environment variable helpers and graceful HTTP server lifecycle handling.

```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/v-e-r-n/quikstop"
)

func main() {
	// 1. Environment Variable Helpers
	// ReqEnv returns an error if the variable is unset or empty
	jwtSecret, err := quikstop.ReqEnv("JWT_SECRET")
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// GetEnv returns a fallback default if the variable is unset
	dbPath := quikstop.GetEnv("DATABASE_PATH", "data/app.db")

	// ResolveAddr constructs host:port checking multiple port fallback keys
	addr := quikstop.ResolveAddr("HOST", []string{"PORT", "HTTP_PORT"}, "8080")

	// 2. Graceful Shutdown Server
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Listens in background and intercepts SIGINT/SIGTERM to drain connections
	quikstop.ListenAndServeGracefully(srv, 5*time.Second)
}
```

---

## 2. Database (`quikstop/db`)

Auto-detects database dialect (SQLite vs PostgreSQL) from connection string scheme, applies SQLite performance tuning (WAL, foreign keys), rebinds `?` placeholders to PostgreSQL `$1, $2` parameters dynamically, and embeds Goose migrations.

```go
import (
	"context"
	"embed"

	"github.com/v-e-r-n/quikstop/db"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

database, err := db.Connect("postgres://user:pass@localhost:5432/app")
// Or: db.Connect("sqlite://data/app.db") or db.Connect("data/app.db")

// Run embedded goose migrations
err = database.Migrate(migrationFS, "migrations")

// Query rebinding: write SQL with '?' for both SQLite and Postgres
query := database.Rebind("SELECT id FROM users WHERE email = ? AND org_id = ?")
row := database.QueryRowContext(ctx, query, email, orgID)
```

---

## 3. JWT Authentication (`quikstop/jwt`)

Generates, verifies, and extracts user/scope IDs from HS256 JWT tokens. The middleware automatically supports both standard `Authorization: Bearer <token>` headers and `?token=` query parameters (ideal for EventSource SSE / WebSocket transports).

```go
import (
	"time"

	"github.com/v-e-r-n/quikstop/jwt"
)

secret := []byte("your-secret-key")

// Issue a token
tokenStr, err := jwt.Generate(userID, secret, 24*time.Hour, "optional-org-id")

// Protect routes
authMiddleware := jwt.Middleware(secret,
	jwt.WithScopeHeaders("X-Organization-Id", "X-Squad-Id"),
)

mux.Handle("GET /api/v1/profile", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	userID := jwt.UserIDFromContext(r.Context())
	orgID := jwt.ScopeIDFromContext(r.Context())
})))
```

---

## 4. Real-time Events & SSE (`quikstop/events`)

Lightweight pub/sub bus and SSE connection dispatcher with targeted recipient resolution.

```go
import (
	"context"
	"time"

	"github.com/v-e-r-n/quikstop/events"
)

bus := events.NewEventBus()
sm := events.NewStreamManager()

// Implement recipient resolution for your domain
type MyResolver struct{}
func (r *MyResolver) ResolveClientIDs(ctx context.Context, ev events.Event) ([]string, error) {
	return []string{"user-123"}, nil
}

dispatcher := events.NewDispatcher(bus, sm, &MyResolver{})
dispatcher.Start(context.Background())

// Expose SSE endpoint
mux.Handle("GET /api/v1/events", events.NewSSEHandler(sm, events.SSEConfig{
	HeartbeatInterval: 25 * time.Second,
	ContextExtractor: func(r *http.Request) (events.ClientSession, error) {
		return events.ClientSession{
			ClientID: jwt.UserIDFromContext(r.Context()),
			ScopeID:  jwt.ScopeIDFromContext(r.Context()),
		}, nil
	},
}))

// Publish event anywhere in your app
bus.Publish(events.NewEvent(
	events.EventType("TASK_UPDATED"),
	events.EventPayload{ResourceID: "task-456"},
	events.EventMeta{ScopeID: "org-1"},
))
```

---

## 5. Rate Limiting (`quikstop/limiter`)

Token-bucket IP rate limiter middleware with proxy-aware IP resolution (`CF-Connecting-IP`, `True-Client-IP`, `X-Forwarded-For`, `X-Real-IP`, and socket fallback) and periodic inactive IP cleanup.

```go
import "github.com/v-e-r-n/quikstop/limiter"

// 10 requests/second with a burst of 20
rateLimiter := limiter.New(10.0, 20.0, limiter.WithOnLimit(func(ip string, r *http.Request) {
	log.Printf("Rate limit exceeded for IP %s on %s", ip, r.URL.Path)
}))

handler := limiter.Handler(rateLimiter)(mux)
```

---

## 6. Passwordless OTP (`quikstop/otp`)

Manages OTP request, secure numeric code generation, expiration TTL, attempt count lockouts, and HTTP endpoints.

```go
import (
	"time"

	"github.com/v-e-r-n/quikstop/otp"
)

store := otp.NewMemoryStore()
deliverer := otp.NewConsoleDeliverer(nil) // Or adapt to mcfeely
keeper := otp.New(store, deliverer, otp.Config{
	CodeLength: 6,
	TTL:        15 * time.Minute,
	MaxRetries: 3,
})

handler := otp.NewHandler(keeper, func(w http.ResponseWriter, r *http.Request, recipient string) {
	// Called on successful verification: issue session / JWT
})

mux.HandleFunc("POST /auth/otp/request", handler.Knock)
mux.HandleFunc("POST /auth/otp/verify", handler.Verify)
```

---

## 7. Email Delivery (`quikstop/mcfeely`)

Speedy delivery interface with standard Console and SMTP adapters.

```go
import "github.com/v-e-r-n/quikstop/mcfeely"

var mailer mcfeely.McFeely

if isDev {
	mailer = mcfeely.NewConsoleMcFeely()
} else {
	mailer = mcfeely.NewSmtpMcFeely("smtp.example.com", "587", "user", "pass", "no-reply@example.com")
}

_ = mailer.Send(ctx, "user@example.com", "Welcome!", "Your account is ready.")
```

---

## 8. CORS (`quikstop/cors`)

Preflight `OPTIONS` and cross-origin header handler.

```go
import "github.com/v-e-r-n/quikstop/cors"

corsMiddleware := cors.Handler(
	cors.WithOrigins("https://app.example.com"),
	cors.WithMethods("GET", "POST", "PUT", "DELETE"),
)

http.ListenAndServe(":8080", corsMiddleware(mux))
```

---

## 9. HTTP Utilities (`quikstop/httputil`)

Standard JSON response writers and error formatters.

```go
import (
	"net/http"

	"github.com/v-e-r-n/quikstop/httputil"
)

func HandleUser(w http.ResponseWriter, r *http.Request) {
	var body CreateUserRequest
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_body", "Failed to parse JSON")
		return
	}

	user, err := CreateUser(body)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	httputil.JSON(w, http.StatusCreated, user)
}
```

---

## Running Tests

```bash
go test -v ./...
```
