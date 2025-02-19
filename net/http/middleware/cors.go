package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"

	keelhttp "github.com/foomo/keel/net/http"
)

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           int
}

var DefaultCORSConfig = CORSConfig{
	AllowOrigins: []string{"*"},
	AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
}

// CORS returns a http middleware with default configuration
func CORS() Middleware {
	return CORSWithConfig(DefaultCORSConfig)
}

// CORSWithConfig returns a http middleware
func CORSWithConfig(config CORSConfig) Middleware {
	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = DefaultCORSConfig.AllowOrigins
	}
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = DefaultCORSConfig.AllowMethods
	}

	allowOriginPatterns := make([]string, len(config.AllowOrigins))
	for i, origin := range config.AllowOrigins {
		pattern := regexp.QuoteMeta(origin)
		pattern = strings.ReplaceAll(pattern, "\\*", ".*")
		pattern = strings.ReplaceAll(pattern, "\\?", ".")
		pattern = "^" + pattern + "$"
		allowOriginPatterns[i] = pattern
	}

	allowMethods := strings.Join(config.AllowMethods, ",")
	allowHeaders := strings.Join(config.AllowHeaders, ",")
	exposeHeaders := strings.Join(config.ExposeHeaders, ",")
	maxAge := strconv.Itoa(config.MaxAge)

	return func(l *zap.Logger, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get(keelhttp.HeaderOrigin)
			allowOrigin := ""

			preflight := r.Method == http.MethodOptions
			w.Header().Add(keelhttp.HeaderVary, keelhttp.HeaderOrigin)

			// No Origin provided
			if origin == "" {
				if !preflight {
					next.ServeHTTP(w, r)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Check allowed origins
			for _, value := range config.AllowOrigins {
				if value == "*" && config.AllowCredentials {
					allowOrigin = origin
					break
				}
				if value == "*" || value == origin {
					allowOrigin = value
					break
				}
				if matchSubdomain(origin, value) {
					allowOrigin = origin
					break
				}
			}

			// Check allowed origin patterns
			for _, re := range allowOriginPatterns {
				if allowOrigin == "" {
					index := strings.Index(origin, "://")
					if index == -1 {
						continue
					}

					if len(origin[index+3:]) > 253 {
						break
					}

					if match, _ := regexp.MatchString(re, origin); match {
						allowOrigin = origin
						break
					}
				}
			}

			// Origin not allowed
			if allowOrigin == "" && !preflight {
				next.ServeHTTP(w, r)
				return
			} else if allowOrigin == "" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Simple request
			if !preflight {
				r.Header.Set(keelhttp.HeaderAccessControlAllowOrigin, allowOrigin)
				if config.AllowCredentials {
					r.Header.Set(keelhttp.HeaderAccessControlAllowCredentials, "true")
				}
				if exposeHeaders != "" {
					r.Header.Set(keelhttp.HeaderAccessControlExposeHeaders, exposeHeaders)
				}
				next.ServeHTTP(w, r)
				return
			}

			// Preflight request
			r.Header.Add(keelhttp.HeaderVary, keelhttp.HeaderAccessControlRequestMethod)
			r.Header.Add(keelhttp.HeaderVary, keelhttp.HeaderAccessControlRequestHeaders)
			r.Header.Set(keelhttp.HeaderAccessControlAllowOrigin, allowOrigin)
			r.Header.Set(keelhttp.HeaderAccessControlAllowMethods, allowMethods)
			if config.AllowCredentials {
				r.Header.Set(keelhttp.HeaderAccessControlAllowCredentials, "true")
			}
			if allowHeaders != "" {
				r.Header.Set(keelhttp.HeaderAccessControlAllowHeaders, allowHeaders)
			} else if h := r.Header.Get(keelhttp.HeaderAccessControlRequestHeaders); h != "" {
				r.Header.Set(keelhttp.HeaderAccessControlAllowHeaders, h)
			}
			if config.MaxAge > 0 {
				r.Header.Set(keelhttp.HeaderAccessControlMaxAge, maxAge)
			}
			w.WriteHeader(http.StatusNoContent)
		})
	}
}

func matchScheme(domain, pattern string) bool {
	didx := strings.Index(domain, ":")
	pidx := strings.Index(pattern, ":")
	return didx != -1 && pidx != -1 && domain[:didx] == pattern[:pidx]
}

// matchSubdomain compares authority with wildcard
func matchSubdomain(domain, pattern string) bool {
	if !matchScheme(domain, pattern) {
		return false
	}
	didx := strings.Index(domain, "://")
	pidx := strings.Index(pattern, "://")
	if didx == -1 || pidx == -1 {
		return false
	}
	domAuth := domain[didx+3:]
	// to avoid long loop by invalid long domain
	if len(domAuth) > 253 {
		return false
	}
	patAuth := pattern[pidx+3:]

	domComp := strings.Split(domAuth, ".")
	patComp := strings.Split(patAuth, ".")
	for i := len(domComp)/2 - 1; i >= 0; i-- {
		opp := len(domComp) - 1 - i
		domComp[i], domComp[opp] = domComp[opp], domComp[i]
	}
	for i := len(patComp)/2 - 1; i >= 0; i-- {
		opp := len(patComp) - 1 - i
		patComp[i], patComp[opp] = patComp[opp], patComp[i]
	}

	for i, v := range domComp {
		if len(patComp) <= i {
			return false
		}
		p := patComp[i]
		if p == "*" {
			return true
		}
		if p != v {
			return false
		}
	}
	return false
}
