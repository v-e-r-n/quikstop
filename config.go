package quikstop

import "os"

// EnvString retrieves an environment variable or returns the default value.
func EnvString(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
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
