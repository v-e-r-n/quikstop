package limiter

import (
	"net"
	"net/http"
	"strings"
)

// GetClientIP extracts the client IP address from the request headers using prioritized fallback.
func GetClientIP(r *http.Request) string {
	// 1. Cloudflare CDN Headers
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		if ip := net.ParseIP(strings.TrimSpace(cfIP)); ip != nil {
			return ip.String()
		}
	}
	if trueIP := r.Header.Get("True-Client-IP"); trueIP != "" {
		if ip := net.ParseIP(strings.TrimSpace(trueIP)); ip != nil {
			return ip.String()
		}
	}

	// 2. X-Forwarded-For (XFF) multi-hop header (parse right-to-left for first public IP)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		var firstPrivate string
		for i := len(parts) - 1; i >= 0; i-- {
			part := strings.TrimSpace(parts[i])
			if host, _, err := net.SplitHostPort(part); err == nil {
				part = host
			}
			ip := net.ParseIP(part)
			if ip != nil {
				if !isPrivateIP(ip) {
					return ip.String()
				}
				if firstPrivate == "" {
					firstPrivate = ip.String()
				}
			}
		}
		if firstPrivate != "" {
			return firstPrivate
		}
	}

	// 3. X-Real-IP
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		if ip := net.ParseIP(strings.TrimSpace(realIP)); ip != nil {
			return ip.String()
		}
	}

	// 4. Fallback to TCP Socket RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		if ip := net.ParseIP(host); ip != nil {
			return ip.String()
		}
		return host
	}
	return r.RemoteAddr
}

// isPrivateIP checks if an IP belongs to private/loopback/link-local ranges.
func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() {
		return true
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 10 ||
			(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) ||
			(ip4[0] == 192 && ip4[1] == 168)
	}
	return len(ip) == 16 && ip[0] == 0xfd
}
