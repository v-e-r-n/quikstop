package quikstop

import (
	"fmt"
	"os"
)

// GetEnv retrieves an environment variable or returns the default value.
func GetEnv(key, defaultVal string) string {
	if val, err := ReqEnv(key); err == nil {
		return val
	}
	return defaultVal
}

// ReqEnv retrieves an environment variable or returns an error if it is blank.
func ReqEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("required environment variable %s is empty", key)
	}
	return val, nil
}

// ResolveAddr builds a host:port TCP address from the specified environment keys.
func ResolveAddr(hostEnv string, portEnvs []string, defaultPort string) string {
	host := os.Getenv(hostEnv)
	port := ""
	for _, portEnv := range portEnvs {
		if val := os.Getenv(portEnv); val != "" {
			port = val
			break
		}
	}
	if port == "" {
		port = defaultPort
	}
	return host + ":" + port
}
