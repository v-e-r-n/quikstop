package cors

import "net/http"

// Config defines the configuration for the CORS middleware.
type Config struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// Option configures CORS settings.
type Option func(*Config)

// WithOrigins sets custom allowed origins.
func WithOrigins(origins ...string) Option {
	return func(c *Config) {
		c.AllowedOrigins = origins
	}
}

// WithMethods sets custom allowed HTTP methods.
func WithMethods(methods ...string) Option {
	return func(c *Config) {
		c.AllowedMethods = methods
	}
}

// WithHeaders sets custom allowed headers.
func WithHeaders(headers ...string) Option {
	return func(c *Config) {
		c.AllowedHeaders = headers
	}
}

// DefaultConfig returns standard open CORS defaults suitable for development and API services.
func DefaultConfig() Config {
	return Config{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Cache-Control",
			"X-Requested-With",
			"X-Organization-Id",
			"X-Squad-Id",
		},
	}
}

// Handler returns a standard CORS middleware wrapper.
func Handler(opts ...Option) func(http.Handler) http.Handler {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	origins := join(cfg.AllowedOrigins, ", ")
	methods := join(cfg.AllowedMethods, ", ")
	headers := join(cfg.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origins)
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func join(items []string, sep string) string {
	if len(items) == 0 {
		return ""
	}
	result := items[0]
	for _, item := range items[1:] {
		result += sep + item
	}
	return result
}
