package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewImageCatalogProxy(targetBaseURL string) gin.HandlerFunc {
	target, err := url.Parse(targetBaseURL)
	if err != nil {
		panic("Invalid target URL for image-catalog-service")
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			originalPath := req.URL.Path
			// /api/v1/image-catalog/... -> /api/v1/images/...
			trimmed := strings.TrimPrefix(originalPath, "/api/v1/image-catalog")

			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = trimmed
			req.Host = target.Host

			// Context'ten gelen user_id ve role header olarak ekle
			if userID := req.Context().Value("user_id"); userID != nil {
				req.Header.Set("X-User-ID", userID.(string))
			}
			if role := req.Context().Value("user_role"); role != nil {
				req.Header.Set("X-User-Role", role.(string))
			}
		},
	}

	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request.WithContext(c.Request.Context()))
	}
}
