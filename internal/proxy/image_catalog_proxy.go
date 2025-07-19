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
)

func NewImageCatalogProxy(targetBaseURL string) gin.HandlerFunc {
	target, err := url.Parse(targetBaseURL)
	if err != nil {
		panic("Invalid target URL for image-catalog-service")
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			originalPath := req.URL.Path
			log.Printf("Original path: %s", originalPath)

			// Remove the /api/v1/image-catalog prefix and prepend /api/v1
			trimmed := strings.TrimPrefix(originalPath, "/api/v1/image-catalog")
			if trimmed == "" {
				trimmed = "/"
			}
			newPath := "/api/v1" + trimmed

			log.Printf("Proxying to: %s%s", target.String(), newPath)

			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = newPath
			req.Host = target.Host

			// Context'ten gelen user_id ve role header olarak ekle
			if userID := req.Context().Value("user_id"); userID != nil {
				req.Header.Set("X-User-ID", userID.(string))
			}
			if role := req.Context().Value("user_role"); role != nil {
				req.Header.Set("X-User-Role", role.(string))
			}
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error: %v", err)
			http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
		},
	}

	return func(c *gin.Context) {
		// Add user context for headers
		if user, exists := c.Get("user"); exists {
			if u, ok := user.(*models.User); ok {
				req := c.Request.WithContext(c.Request.Context())
				req = req.WithContext(context.WithValue(req.Context(), "user_id", u.UID))
				req = req.WithContext(context.WithValue(req.Context(), "user_role", string(u.Role)))
				c.Request = req
			}
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
