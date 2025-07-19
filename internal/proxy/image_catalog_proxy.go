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
	"github.com/histopathai/auth-service/internal/models"
	"github.com/histopathai/auth-service/internal/service"
)

func NewImageCatalogProxy(targetBaseURL string, authService service.AuthService, sessionService *service.ImageSessionService) gin.HandlerFunc {
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
		var user *models.User

		// 🔧 ÖNCE Bearer token kontrol et (API endpoints için)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				bearerToken := parts[1]
				log.Printf("🔍 Bearer token found - verifying...")

				// AuthService ile token'ı verify et
				authService := &service.AuthServiceImpl{} // Burayı kendi authService instance'ınıza göre ayarlayın
				verifiedUser, err := authService.VerifyToken(c.Request.Context(), bearerToken)
				if err == nil && verifiedUser != nil {
					user = verifiedUser
					log.Printf("✅ Bearer token verified for user: %s", user.UID)
				} else {
					log.Printf("❌ Bearer token verification failed: %v", err)
				}
			}
		}

		// 🔧 Eğer Bearer token yoksa veya geçersizse, session kontrol et
		if user == nil {
			sessionID := c.Query("session")
			if sessionID != "" {
				log.Printf("🔍 Session ID found: %s", sessionID[:8]+"...")

				session, valid := sessionService.ValidateSession(sessionID)
				if valid && session != nil {
					// Session'dan user'ı al - burada authService.GetUser kullanmanız gerekebilir
					log.Printf("✅ Session validated for user: %s", session.UserID)

					// User objesini oluştur (basit yaklaşım)
					user = &models.User{
						UID:    session.UserID,
						Role:   models.UserRole(session.Role),
						Status: models.StatusActive, // Session varsa aktif kabul et
					}

					// Auto-extend session
					if session.RequestCount%50 == 0 {
						sessionService.ExtendSession(sessionID)
						log.Printf("🔄 Session auto-extended")
					}
				} else {
					log.Printf("❌ Invalid session: %s", sessionID[:8]+"...")
				}
			}
		}

		// 🔧 Hiçbir auth yoksa hata
		if user == nil {
			log.Printf("❌ No valid authentication found")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication_required",
				"message": "Valid Bearer token or session required",
			})
			return
		}

		// User status kontrolü
		if user.Status != models.StatusActive {
			log.Printf("❌ User not active: %s", user.Status)
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "account_inactive",
				"message": "Account not active",
			})
			return
		}

		// Context'e user bilgilerini ekle
		ctx := context.WithValue(c.Request.Context(), "user_id", user.UID)
		ctx = context.WithValue(ctx, "user_role", string(user.Role))
		c.Request = c.Request.WithContext(ctx)

		// Performance logging
		defer func() {
			duration := time.Since(start)
			if duration > 100*time.Millisecond {
				log.Printf("⚠️ Slow request: %s took %v", c.Request.URL.Path, duration)
			}
		}()

		log.Printf("🔍 Proxy: %s %s for user: %s", c.Request.Method, c.Request.URL.Path, user.UID)
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
