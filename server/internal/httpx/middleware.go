package httpx

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/mddanish004/jamzy/server/internal/auth"
	"github.com/mddanish004/jamzy/server/internal/config"
	"github.com/mddanish004/jamzy/server/internal/logging"
	"github.com/mddanish004/jamzy/server/internal/ratelimit"
)

const (
	corsAllowMethods = "GET,HEAD,PUT,PATCH,POST,DELETE"
	bearerPrefix     = "Bearer "
)

type Middleware struct {
	Log         logging.Logger
	Verifier    auth.Verifier
	Limiter     *ratelimit.Limiter
	CORSOrigins []string
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

func (s *statusWriter) Write(payload []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(payload)
}

func (s *statusWriter) Unwrap() http.ResponseWriter {
	return s.ResponseWriter
}

func (s *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := s.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

func (s *statusWriter) Flush() {
	if flusher, ok := s.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (m *Middleware) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}
			m.Log.Error("Unhandled panic", fmt.Sprint(recovered), string(debug.Stack()))
			WriteJSON(w, 500, map[string]string{"message": "Internal server error"})
		}()

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) RequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		writer := &statusWriter{ResponseWriter: w}

		next.ServeHTTP(writer, r)

		status := writer.status
		if status == 0 {
			status = http.StatusOK
		}
		m.Log.Info(fmt.Sprintf("%s %s %d %s", r.Method, r.URL.Path, status, time.Since(started).Round(time.Millisecond)))
	})
}

func (m *Middleware) SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "SAMEORIGIN")
		header.Set("X-DNS-Prefetch-Control", "off")
		header.Set("X-Download-Options", "noopen")
		header.Set("X-Permitted-Cross-Domain-Policies", "none")
		header.Set("X-XSS-Protection", "0")
		header.Set("Origin-Agent-Cluster", "?1")
		header.Set("Referrer-Policy", "no-referrer")
		header.Set("Strict-Transport-Security", "max-age=15552000; includeSubDomains")
		header.Set("Cross-Origin-Opener-Policy", "same-origin")

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) CORS(next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(m.CORSOrigins))
	for _, origin := range m.CORSOrigins {
		allowed[origin] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		preflight := r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""

		if origin != "" {
			if _, ok := allowed[origin]; ok {
				header := w.Header()
				header.Set("Access-Control-Allow-Origin", origin)
				header.Add("Vary", "Origin")
				header.Set("Access-Control-Allow-Credentials", "true")

				if preflight {
					header.Set("Access-Control-Allow-Methods", corsAllowMethods)
					header.Add("Vary", "Access-Control-Request-Headers")
					if requested := r.Header.Get("Access-Control-Request-Headers"); requested != "" {
						header.Set("Access-Control-Allow-Headers", requested)
					}
				}
			}

			if preflight {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) NotFound(w http.ResponseWriter, _ *http.Request) {
	WriteJSON(w, 404, map[string]string{"message": "Not found"})
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			WriteUnauthorized(w)
			return
		}

		clerkUserID, err := m.Verifier.Verify(r.Context(), token)
		if err != nil {
			m.Log.Warn("Authentication rejected", err)
			WriteUnauthorized(w)
			return
		}

		identity := auth.Identity{ClerkUserID: clerkUserID}
		next.ServeHTTP(w, r.WithContext(auth.WithIdentity(r.Context(), identity)))
	})
}

func (m *Middleware) RateLimit(bucket string, rule config.RateLimitRule) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			result := m.Limiter.Consume(bucket+":"+requestKey(r), rule)
			if result.Allowed {
				next.ServeHTTP(w, r)
				return
			}

			retryAfter := ratelimit.RetryAfterSeconds(result.RetryAfterMs)
			m.Log.Warn("Rate limit exceeded", bucket, r.Method)
			w.Header().Set("Retry-After", fmt.Sprint(retryAfter))
			WriteJSON(w, 429, map[string]string{
				"message": fmt.Sprintf("Too many requests. Try again in %ds.", retryAfter),
			})
		})
	}
}

func (m *Middleware) Route(bucket string, rule config.RateLimitRule, fn Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.RequireAuth(m.RateLimit(bucket, rule)(http.HandlerFunc(m.Handle(fn)))).ServeHTTP(w, r)
	}
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" || !strings.HasPrefix(header, bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
}

func requestKey(r *http.Request) string {
	if clerkUserID, err := auth.ClerkUserID(r.Context()); err == nil {
		return "user:" + clerkUserID
	}
	return "ip:" + remoteAddress(r)
}

func remoteAddress(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
