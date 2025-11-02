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

func NewMainServiceProxy(targetBaseURL string, authService service.AuthService, sessionService *service.ImageSessionService) gin.HandlerFunc {
	target, err := url.Parse(targetBaseURL)
	if err != nil {
		log.Printf("❌ HATA: Hedef URL geçersiz: %v", err)
		panic(fmt.Sprintf("Invalid target URL for main-service: %v", err))
	}

	if !strings.HasSuffix(targetBaseURL, "/") {
		targetBaseURL = targetBaseURL + "/"
		log.Printf("🔧 Target URL Fixed: %s", targetBaseURL)
	}

	log.Printf("🔧 Main Service Proxy hedefi: %s", targetBaseURL)

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			originalPath := req.URL.Path
			originalMethod := req.Method

			log.Printf("📥 Proxy request: %s %s", originalMethod, originalPath)

			trimmed := strings.TrimPrefix(originalPath, "/api/v1/main")
			if trimmed == "" {
				trimmed = "/"
			}
			newPath := "/api/v1" + trimmed

			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = newPath
			req.Host = target.Host

			if req.URL.RawQuery != "" {
				log.Printf("🔍 Query parameters: %s", req.URL.RawQuery)
			}

			sessionID := req.URL.Query().Get("session")
			if sessionID == "" {
				log.Printf("⚠️ Session ID not found! URL: %s", originalPath)
			} else {
				req.Header.Set("X-Image-Session-ID", sessionID)
				values := req.URL.Query()
				values.Del("session")
				req.URL.RawQuery = values.Encode()
				log.Printf("✅ Session ID found and added to header: %s...", sessionID[:8])
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

			log.Printf("📤 Target Proxy: %s://%s%s", req.URL.Scheme, req.URL.Host, req.URL.Path)
		},
		ModifyResponse: func(resp *http.Response) error {
			statusCode := resp.StatusCode
			requestURL := resp.Request.URL.String()

			if statusCode >= 200 && statusCode < 300 {
				log.Printf("✅ Proxy response: %d for %s", statusCode, requestURL)

				if strings.Contains(resp.Request.URL.Path, "/proxy/") {
					resp.Header.Set("Cache-Control", "public, max-age=3600")
					resp.Header.Set("ETag", `"`+resp.Request.URL.Path+`"`)
				}
				return nil
			}

			log.Printf("❌ Proxy error response: %d for %s", statusCode, requestURL)

			var body []byte
			if resp.Body != nil {
				body, _ = io.ReadAll(resp.Body)
				resp.Body = io.NopCloser(bytes.NewBuffer(body))

				maxLength := 500
				logContent := string(body)
				if len(logContent) > maxLength {
					logContent = logContent[:maxLength] + "..."
				}
				log.Printf("❌ Error content: %s", logContent)
			}

			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("❌ Proxy request failed: %v, URL: %s", err, r.URL.String())

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
				"message":    "The main service is currently unavailable. Please try again later.",
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
				log.Printf("✅ Session verified user: %s", session.UserID)

				user = &models.User{
					UID:    session.UserID,
					Role:   models.UserRole(session.Role),
					Status: models.StatusActive,
				}

				// Auto-extend session
				if session.RequestCount%50 == 0 {
					sessionService.ExtendSession(sessionID)
					log.Printf("🔄 Session automatically extended")
				}
			} else {
				log.Printf("❌ Invalid session: %s", sessionID[:8])
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "invalid_session",
					"message": "Invalid or expired session ID",
				})
				return
			}
		}

		if user == nil {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					bearerToken := parts[1]
					log.Printf("🔍 Bearer token found - verifying...")

					verifiedUser, err := authService.VerifyToken(c.Request.Context(), bearerToken)
					if err == nil && verifiedUser != nil {
						user = verifiedUser
						log.Printf("✅ Bearer token verified, user: %s", user.UID)
					} else {
						log.Printf("❌ Bearer token verification failed: %v", err)
					}
				}
			}
		}

		if user == nil {
			log.Printf("❌ Valid authentication not found")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication_required",
				"message": "Valid Bearer token or session required",
			})
			return
		}

		// 4. User status
		if user.Status != models.StatusActive {
			log.Printf("❌ User Not Active: %s", user.Status)
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "account_inactive",
				"message": "Account is inactive. Please contact support.",
			})
			return
		}

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

		log.Printf("🔍 Proxy request is being forwarded: %s %s user: %s", c.Request.Method, c.Request.URL.Path, user.UID)
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
