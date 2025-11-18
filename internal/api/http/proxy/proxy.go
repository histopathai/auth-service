package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/internal/domain/model"
	"github.com/histopathai/auth-service/internal/service"
	"github.com/histopathai/auth-service/pkg/config"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

type MainServiceProxy struct {
	targetURL      *url.URL
	proxy          *httputil.ReverseProxy
	authService    *service.AuthService
	sessionService *service.SessionService
	logger         *slog.Logger
	config         *config.Config
	tokenSource    oauth2.TokenSource
}

func NewMainServiceProxy(
	targetBaseURL string,
	authService *service.AuthService,
	sessionService *service.SessionService,
	config *config.Config,
	logger *slog.Logger,
) (*MainServiceProxy, error) {
	target, err := url.Parse(targetBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL: %w", err)
	}

	if !strings.HasSuffix(targetBaseURL, "/") {
		targetBaseURL = targetBaseURL + "/"
		target, _ = url.Parse(targetBaseURL)
	}

	ts, err := idtoken.NewTokenSource(context.Background(), targetBaseURL)
	if err != nil {
		// Local development'ta hata vermemesi için loglayıp geçebilirsiniz veya mocklayabilirsiniz
		logger.Warn("Failed to create ID token source (ignore if local)", "error", err)
	}
	msp := &MainServiceProxy{
		targetURL:      target,
		authService:    authService,
		sessionService: sessionService,
		config:         config,
		logger:         logger,
		tokenSource:    ts,
	}

	msp.proxy = &httputil.ReverseProxy{
		Director:       msp.director,
		ModifyResponse: msp.modifyResponse,
		ErrorHandler:   msp.errorHandler,
	}

	logger.Info("Main Service Proxy initialized",
		"target", targetBaseURL,
	)

	return msp, nil
}

func (msp *MainServiceProxy) director(req *http.Request) {
	originalPath := req.URL.Path
	originalMethod := req.Method

	msp.logger.Debug("Proxying request",
		"method", originalMethod,
		"path", originalPath,
		"query", req.URL.RawQuery,
	)

	trimmed := strings.TrimPrefix(originalPath, "/api/v1/proxy")
	if trimmed == "" {
		trimmed = "/"
	}

	newPath := "/api/v1" + trimmed
	req.URL.Scheme = msp.targetURL.Scheme
	req.URL.Host = msp.targetURL.Host
	req.URL.Path = newPath
	req.Host = msp.targetURL.Host

	if msp.tokenSource != nil {
		token, err := msp.tokenSource.Token()
		if err == nil {
			req.Header.Set("Authorization", "Bearer "+token.AccessToken)
		} else {
			msp.logger.Error("Failed to refresh ID token", "error", err)
		}
	}

	// Move Session Token to Header
	if sessionID := req.URL.Query().Get("session"); sessionID != "" {
		req.Header.Set("X-Session-ID", sessionID)

		// Remove session from query params
		values := req.URL.Query()
		values.Del("session")
		req.URL.RawQuery = values.Encode()

		msp.logger.Debug("Session moved to header",
			"session_id", sessionID[:min(8, len(sessionID))],
		)
	}

	// Move user info to headers from context
	if userID, ok := req.Context().Value("user_id").(string); ok {
		req.Header.Set("X-User-ID", userID)
	}

	if role, ok := req.Context().Value("user_role").(string); ok {
		req.Header.Set("X-User-Role", role)
	}

	msp.logger.Debug("Request proxied",
		"target_url", fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.URL.Host, req.URL.Path),
	)
}

func (msp *MainServiceProxy) modifyResponse(resp *http.Response) error {
	statusCode := resp.StatusCode
	requestURL := resp.Request.URL.String()

	// Remove any CORS headers from backend - we'll handle them in the Handler
	resp.Header.Del("Access-Control-Allow-Origin")
	resp.Header.Del("Access-Control-Allow-Credentials")
	resp.Header.Del("Access-Control-Allow-Methods")
	resp.Header.Del("Access-Control-Allow-Headers")
	resp.Header.Del("Access-Control-Max-Age")

	if statusCode >= 200 && statusCode < 300 {
		msp.logger.Debug("Proxy response",
			"status", statusCode,
			"url", requestURL,
		)

		// Cache headers for image/tile endpoints
		if strings.Contains(resp.Request.URL.Path, "/tiles/") ||
			strings.Contains(resp.Request.URL.Path, "/images/") {
			resp.Header.Set("Cache-Control", "public, max-age=3600")
			resp.Header.Set("ETag", fmt.Sprintf(`"%s"`, resp.Request.URL.Path))
		}

		return nil
	}

	// Error response handling
	msp.logger.Warn("Proxy error response",
		"status", statusCode,
		"url", requestURL,
	)

	// Read and log error body
	if resp.Body != nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(body))

		if len(body) > 0 && len(body) < 1000 {
			msp.logger.Warn("Error response body",
				"body", string(body),
			)
		}
	}

	return nil
}

func (msp *MainServiceProxy) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	msp.logger.Error("Proxy request failed",
		"error", err,
		"url", r.URL.String(),
		"method", r.Method,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)

	errorType := "connection_error"
	if strings.Contains(err.Error(), "timeout") {
		errorType = "timeout_error"
	} else if strings.Contains(err.Error(), "connection refused") {
		errorType = "connection_refused"
	}

	errorResponse := map[string]interface{}{
		"error":   "service_unavailable",
		"message": "Main service is temporarily unavailable",
		"details": errorType,
	}

	json.NewEncoder(w).Encode(errorResponse)
}

func (msp *MainServiceProxy) setCORSHeaders(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")

	if origin != msp.config.AllowedOrigin {
		return
	}

	c.Writer.Header().Set("Access-Control-Allow-Origin", msp.config.AllowedOrigin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Session-ID, Cookie")
	c.Writer.Header().Set("Access-Control-Max-Age", "3600")
	c.Writer.Header().Set("Access-Control-Expose-Headers", "Set-Cookie")
}

// Handler returns the Gin handler function
func (msp *MainServiceProxy) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Set CORS headers first
		msp.setCORSHeaders(c)

		// Handle OPTIONS requests (preflight)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		// Authenticate request
		user, err := msp.authenticateRequest(c)
		if err != nil {
			msp.handleAuthError(c, err)
			return
		}

		// Check user status
		if user.Status != model.StatusActive {
			msp.logger.Warn("Inactive user attempted to access proxy",
				"user_id", user.UserID,
				"status", user.Status,
			)
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "account_inactive",
				"message": "Account is not active",
			})
			return
		}

		// Add user info to context
		ctx := context.WithValue(c.Request.Context(), "user_id", user.UserID)
		ctx = context.WithValue(ctx, "user_role", string(user.Role))
		c.Request = c.Request.WithContext(ctx)

		// Log slow requests
		defer func() {
			duration := time.Since(start)
			if duration > 2*time.Second {
				msp.logger.Warn("Slow proxy request",
					"duration", duration,
					"path", c.Request.URL.Path,
					"user_id", user.UserID,
				)
			}
		}()

		msp.logger.Info("Proxying authenticated request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"user_id", user.UserID,
			"role", user.Role,
		)

		// Proxy the request
		msp.proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func (msp *MainServiceProxy) authenticateRequest(c *gin.Context) (*model.User, error) {
	// 1. Try session authentication first (highest priority)
	if sessionID, err := c.Cookie("session_id"); err == nil && sessionID != "" {
		msp.logger.Debug("Attempting session cookie authentication",
			"session_id", sessionID[:min(8, len(sessionID))],
		)

		session, err := msp.sessionService.ValidateAndExtend(c.Request.Context(), sessionID)
		if err == nil && session != nil {
			user, err := msp.authService.GetUserByUserID(c.Request.Context(), session.UserID)
			if err == nil {
				msp.logger.Debug("Session cookie authentication successful",
					"user_id", user.UserID,
				)
				return user, nil
			}
		}

		msp.logger.Warn("Session cookie authentication failed",
			"session_id", sessionID[:min(8, len(sessionID))],
			"error", err,
		)
	}

	// 2. Try bearer token authentication
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			bearerToken := parts[1]

			msp.logger.Debug("Attempting bearer token authentication")

			user, err := msp.authService.VerifyToken(c.Request.Context(), bearerToken)
			if err == nil && user != nil {
				msp.logger.Debug("Bearer token authentication successful",
					"user_id", user.UserID,
				)
				return user, nil
			}

			msp.logger.Warn("Bearer token authentication failed",
				"error", err,
			)
		}
	}

	return nil, fmt.Errorf("no valid authentication found")
}

func (msp *MainServiceProxy) handleAuthError(c *gin.Context, err error) {
	msp.logger.Warn("Authentication failed for proxy request",
		"error", err,
		"path", c.Request.URL.Path,
	)

	c.JSON(http.StatusUnauthorized, gin.H{
		"error":   "authentication_required",
		"message": "Valid Bearer token or session required",
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
