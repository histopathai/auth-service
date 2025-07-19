// internal/proxy/optimized_image_catalog_proxy.go
package proxy

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/service"
)

func NewImageCatalogProxy(targetBaseURL string, sessionService *service.ImageSessionService) gin.HandlerFunc {
	target, err := url.Parse(targetBaseURL)
	if err != nil {
		panic("Invalid target URL for image-catalog-service")
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			originalPath := req.URL.Path

			// Path transformation
			trimmed := strings.TrimPrefix(originalPath, "/api/v1/image-catalog")
			if trimmed == "" {
				trimmed = "/"
			}
			newPath := "/api/v1" + trimmed

			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = newPath
			req.Host = target.Host

			// Remove session from query params before forwarding
			if req.URL.RawQuery != "" {
				values := req.URL.Query()
				values.Del("session")
				req.URL.RawQuery = values.Encode()
			}

			// Add user headers from context
			if userID := req.Context().Value("user_id"); userID != nil {
				req.Header.Set("X-User-ID", userID.(string))
			}
			if role := req.Context().Value("user_role"); role != nil {
				req.Header.Set("X-User-Role", role.(string))
			}
		},
		ModifyResponse: func(resp *http.Response) error {
			// Cache headers for static assets (tiles, thumbnails)
			if strings.Contains(resp.Request.URL.Path, "/proxy/") {
				resp.Header.Set("Cache-Control", "public, max-age=3600") // 1 saat cache
				resp.Header.Set("ETag", `"`+resp.Request.URL.Path+`"`)
			}
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("❌ Proxy error: %v for URL: %s", err, r.URL.String())
			http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
		},
	}

	return func(c *gin.Context) {
		start := time.Now()

		// Session-based authentication
		sessionID := c.Query("session")
		if sessionID == "" {
			log.Printf("❌ No session ID provided")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "session_required",
				"message": "Session ID required for image access",
			})
			return
		}

		// Validate session (very fast - memory lookup)
		session, valid := sessionService.ValidateSession(sessionID)
		if !valid {
			log.Printf("❌ Invalid session: %s", sessionID)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_session",
				"message": "Session expired or invalid",
			})
			return
		}

		// Auto-extend session if heavily used (smart extension)
		if session.RequestCount%50 == 0 { // Her 50 request'te bir extend et
			sessionService.ExtendSession(sessionID)
			log.Printf("🔄 Session auto-extended: %s", sessionID)
		}

		// Add user context
		ctx := context.WithValue(c.Request.Context(), "user_id", session.UserID)
		ctx = context.WithValue(ctx, "user_role", session.Role)
		c.Request = c.Request.WithContext(ctx)

		// Performance logging
		defer func() {
			duration := time.Since(start)
			if duration > 100*time.Millisecond {
				log.Printf("⚠️ Slow request: %s took %v", c.Request.URL.Path, duration)
			}
		}()

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
