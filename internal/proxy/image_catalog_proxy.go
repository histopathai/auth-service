package proxy

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/service"
)

func NewImageCatalogProxy(targetBaseURL string, authService service.AuthService) gin.HandlerFunc {
	target, err := url.Parse(targetBaseURL)
	if err != nil {
		panic("Invalid target URL for image-catalog-service")
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			originalPath := req.URL.Path
			log.Printf("🔍 Proxy: Original path: %s", originalPath)

			// Remove the /api/v1/image-catalog prefix and prepend /api/v1
			trimmed := strings.TrimPrefix(originalPath, "/api/v1/image-catalog")
			if trimmed == "" {
				trimmed = "/"
			}
			newPath := "/api/v1" + trimmed

			log.Printf("🔍 Proxy: New path: %s", newPath)
			log.Printf("🔍 Proxy: Target URL: %s", target.String())
			log.Printf("🔍 Proxy: Full proxied URL: %s%s", target.String(), newPath)

			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = newPath
			req.Host = target.Host

			// Forward query parameters (but remove token)
			if req.URL.RawQuery != "" {
				log.Printf("🔍 Proxy: Query params: %s", req.URL.RawQuery)
				// Remove token from query params before forwarding
				values := req.URL.Query()
				values.Del("token")
				req.URL.RawQuery = values.Encode()
			}

			// Add user context headers (if available)
			if userID := req.Context().Value("user_id"); userID != nil {
				req.Header.Set("X-User-ID", userID.(string))
				log.Printf("🔍 Proxy: Added X-User-ID header: %s", userID.(string))
			}
			if role := req.Context().Value("user_role"); role != nil {
				req.Header.Set("X-User-Role", role.(string))
				log.Printf("🔍 Proxy: Added X-User-Role header: %s", role.(string))
			}
		},
		ModifyResponse: func(resp *http.Response) error {
			// Log response details
			log.Printf("🔍 Proxy: Response status: %d", resp.StatusCode)
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("❌ Proxy error: %v", err)
			log.Printf("❌ Proxy error for URL: %s", r.URL.String())
			http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
		},
	}

	return func(c *gin.Context) {
		// Check for token in query parameter (for direct requests like DZI, tiles, thumbnails)
		token := c.Query("token")
		var user *models.User

		if token != "" {
			// Verify token from query parameter
			verifiedUser, err := authService.VerifyToken(c.Request.Context(), token)
			if err != nil {
				log.Printf("❌ Token verification failed: %v", err)
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "invalid_token",
					"message": "Token verification failed",
				})
				return
			}
			user = verifiedUser
			log.Printf("✅ Token verified from query param for user: %s", user.UID)
		} else {
			// Try to get user from context (set by middleware)
			if userFromContext, exists := c.Get("user"); exists {
				if u, ok := userFromContext.(*models.User); ok {
					user = u
					log.Printf("✅ User found in context: %s", user.UID)
				}
			}
		}

		// Ensure user is authenticated and active
		if user == nil {
			log.Printf("❌ No authenticated user found")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication_required",
				"message": "Authentication required",
			})
			return
		}

		if user.Status != models.StatusActive {
			log.Printf("❌ User not active: %s", user.Status)
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "account_inactive",
				"message": "Account is not active",
			})
			return
		}

		// Add user context for headers
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "user_id", user.UID)
		ctx = context.WithValue(ctx, "user_role", string(user.Role))
		c.Request = c.Request.WithContext(ctx)

		log.Printf("🔍 Proxy: Processing request: %s %s for user: %s", c.Request.Method, c.Request.URL.Path, user.UID)
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
