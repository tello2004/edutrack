package http

import "net/http"

// CORS middleware configuration.
type CORSConfig struct {
	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	// Use "*" to allow any origin, or specify specific origins.
	AllowedOrigins []string

	// AllowedMethods is a list of methods the client is allowed to use.
	AllowedMethods []string

	// AllowedHeaders is a list of headers the client is allowed to send.
	AllowedHeaders []string

	// ExposedHeaders is a list of headers that are safe to expose to the API.
	ExposedHeaders []string

	// AllowCredentials indicates whether the request can include user credentials.
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached.
	MaxAge string
}

// DefaultCORSConfig returns a permissive CORS configuration suitable for development.
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		ExposedHeaders:   []string{"Content-Length", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           "86400", // 24 hours
	}
}

// withCORS wraps an http.Handler with CORS headers.
func withCORS(config *CORSConfig, next http.Handler) http.Handler {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Set CORS headers.
		if origin != "" {
			// Check if origin is allowed.
			allowed := false
			for _, o := range config.AllowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed {
				if config.AllowedOrigins[0] == "*" && !config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}

				if config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}
		}

		// Handle preflight requests.
		if r.Method == http.MethodOptions {
			// Set preflight headers.
			if len(config.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", joinStrings(config.AllowedMethods))
			}

			if len(config.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", joinStrings(config.AllowedHeaders))
			}

			if len(config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", joinStrings(config.ExposedHeaders))
			}

			if config.MaxAge != "" {
				w.Header().Set("Access-Control-Max-Age", config.MaxAge)
			}

			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Call the next handler.
		next.ServeHTTP(w, r)
	})
}

// joinStrings joins a slice of strings with commas.
func joinStrings(s []string) string {
	if len(s) == 0 {
		return ""
	}

	result := s[0]
	for i := 1; i < len(s); i++ {
		result += ", " + s[i]
	}
	return result
}
