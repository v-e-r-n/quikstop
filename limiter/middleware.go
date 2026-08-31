package limiter

import "net/http"

// Handler returns an HTTP middleware function that rate-limits requests by client IP.
func Handler(l *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetClientIP(r)
			if !l.Allow(ip) {
				if l.onLimit != nil {
					l.onLimit(ip, r)
				}
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
