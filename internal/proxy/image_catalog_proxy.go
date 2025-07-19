// internal/proxy/image_catalog_proxy.go
package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
		log.Printf("❌ HATA: Hedef URL geçersiz: %v", err)
		panic(fmt.Sprintf("Invalid target URL for image-catalog-service: %v", err))
	}

	// URL'nin doğru olduğundan emin ol - sonunda / yoksa ekle
	if !strings.HasSuffix(targetBaseURL, "/") {
		targetBaseURL = targetBaseURL + "/"
		log.Printf("🔧 Hedef URL düzeltildi: %s", targetBaseURL)
	}

	log.Printf("🔧 Image Catalog Proxy hedefi: %s", targetBaseURL)

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			originalPath := req.URL.Path
			originalMethod := req.Method

			// URL ve metodu logla
			log.Printf("📥 Proxy istek: %s %s", originalMethod, originalPath)

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

			// Query parametrelerini logla
			if req.URL.RawQuery != "" {
				log.Printf("🔍 Sorgu parametreleri: %s", req.URL.RawQuery)
			}

			// Session parametresini kontrol et
			sessionID := req.URL.Query().Get("session")
			if sessionID == "" {
				log.Printf("⚠️ İstekte session parametresi bulunamadı! URL: %s", originalPath)
			} else {
				// Session başarıyla alındı, X-Image-Session-ID header'ına ekle ve query'den kaldır
				req.Header.Set("X-Image-Session-ID", sessionID)
				values := req.URL.Query()
				values.Del("session")
				req.URL.RawQuery = values.Encode()
				log.Printf("✅ Session parametresi alındı ve header'a eklendi: %s...", sessionID[:8])
			}

			// User bilgilerini header'lara ekle
			if userID := req.Context().Value("user_id"); userID != nil {
				req.Header.Set("X-User-ID", userID.(string))
				log.Printf("👤 User ID header eklendi: %s", userID.(string))
			}

			if role := req.Context().Value("user_role"); role != nil {
				req.Header.Set("X-User-Role", role.(string))
				log.Printf("🔑 User Role header eklendi: %s", role.(string))
			}

			log.Printf("📤 Proxy hedef: %s://%s%s", req.URL.Scheme, req.URL.Host, req.URL.Path)
		},
		ModifyResponse: func(resp *http.Response) error {
			statusCode := resp.StatusCode
			requestURL := resp.Request.URL.String()

			// Başarılı yanıtlar için sadece bilgi logu
			if statusCode >= 200 && statusCode < 300 {
				log.Printf("✅ Proxy yanıtı: %d for %s", statusCode, requestURL)

				// İmaj proxy istekleri için cache headerları ekle
				if strings.Contains(resp.Request.URL.Path, "/proxy/") {
					resp.Header.Set("Cache-Control", "public, max-age=3600") // 1 saat cache
					resp.Header.Set("ETag", `"`+resp.Request.URL.Path+`"`)
				}
				return nil
			}

			// Hatalı yanıtlar için detaylı loglama
			log.Printf("❌ Proxy hata yanıtı: %d for %s", statusCode, requestURL)

			// Hata içeriğini oku ve logla
			var body []byte
			if resp.Body != nil {
				body, _ = io.ReadAll(resp.Body)
				// Yanıt gövdesini geri koy
				resp.Body = io.NopCloser(bytes.NewBuffer(body))

				// Hata içeriğini logla (ilk 500 karakter)
				maxLength := 500
				logContent := string(body)
				if len(logContent) > maxLength {
					logContent = logContent[:maxLength] + "..."
				}
				log.Printf("❌ Hata içeriği: %s", logContent)
			}

			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("❌ Proxy isteği başarısız: %v, URL: %s", err, r.URL.String())

			// JSON formatında hata döndür
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)

			errorType := "connection_error"
			if strings.Contains(err.Error(), "timeout") {
				errorType = "timeout_error"
			} else if strings.Contains(err.Error(), "connection refused") {
				errorType = "connection_refused"
			}

			errorResponse := map[string]interface{}{
				"error":      "service_unavailable",
				"message":    "Image catalog service geçici olarak ulaşılamıyor",
				"details":    errorType,
				"error_info": err.Error(),
			}

			json.NewEncoder(w).Encode(errorResponse)
		},
	}

	return func(c *gin.Context) {
		start := time.Now()
		var user *models.User

		// CORS headerları ekle
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Image-Session-ID")

		// OPTIONS isteklerini hemen yanıtla
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		// 1. Önce session parametresi kontrolü - en yüksek öncelikli
		sessionID := c.Query("session")
		if sessionID != "" {
			log.Printf("🔍 Session ID bulundu: %s...", sessionID[:8])

			session, valid := sessionService.ValidateSession(sessionID)
			if valid && session != nil {
				log.Printf("✅ Session doğrulandı, kullanıcı: %s", session.UserID)

				// Basit user objesi oluştur
				user = &models.User{
					UID:    session.UserID,
					Role:   models.UserRole(session.Role),
					Status: models.StatusActive, // Session varsa aktif kabul et
				}

				// Auto-extend session
				if session.RequestCount%50 == 0 {
					sessionService.ExtendSession(sessionID)
					log.Printf("🔄 Session otomatik uzatıldı")
				}
			} else {
				log.Printf("❌ Geçersiz session: %s", sessionID[:8])
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "invalid_session",
					"message": "Geçersiz veya süresi dolmuş session",
				})
				return
			}
		}

		// 2. Session yoksa veya geçersizse, Bearer token kontrolü
		if user == nil {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					bearerToken := parts[1]
					log.Printf("🔍 Bearer token bulundu - doğrulanıyor...")

					// AuthService ile token'ı verify et
					verifiedUser, err := authService.VerifyToken(c.Request.Context(), bearerToken)
					if err == nil && verifiedUser != nil {
						user = verifiedUser
						log.Printf("✅ Bearer token doğrulandı, kullanıcı: %s", user.UID)
					} else {
						log.Printf("❌ Bearer token doğrulaması başarısız: %v", err)
					}
				}
			}
		}

		// 3. Hiçbir kimlik doğrulama bulunamadıysa, hata döndür
		if user == nil {
			log.Printf("❌ Geçerli kimlik doğrulama bulunamadı")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication_required",
				"message": "Geçerli Bearer token veya session gerekli",
			})
			return
		}

		// 4. User status kontrolü
		if user.Status != models.StatusActive {
			log.Printf("❌ Kullanıcı aktif değil: %s", user.Status)
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "account_inactive",
				"message": "Hesap aktif değil",
			})
			return
		}

		// 5. Context'e user bilgilerini ekle
		ctx := context.WithValue(c.Request.Context(), "user_id", user.UID)
		ctx = context.WithValue(ctx, "user_role", string(user.Role))
		c.Request = c.Request.WithContext(ctx)

		// Performance logging
		defer func() {
			duration := time.Since(start)
			if duration > 100*time.Millisecond {
				log.Printf("⚠️ Yavaş istek: %s için %v sürdü", c.Request.URL.Path, duration)
			}
		}()

		log.Printf("🔍 Proxy istek yönlendiriliyor: %s %s kullanıcı: %s", c.Request.Method, c.Request.URL.Path, user.UID)
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
