package proxy

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/auth-service/pkg/config"
)

// A fake main-service records what reaches it.
func publicProxy(t *testing.T) (*httptest.Server, *http.Request) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	got := &http.Request{}
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*got = *r.Clone(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(backend.Close)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	msp, err := NewMainServiceProxy(backend.URL, nil, nil, &config.Config{}, logger)
	if err != nil {
		t.Fatal(err)
	}
	msp.tokenSource = nil // no ID token in tests
	engine := gin.New()
	engine.Any("/api/v1/public/blind-tests/*path", msp.PublicHandler())
	front := httptest.NewServer(engine) // ReverseProxy needs a real connection, not a recorder
	t.Cleanup(front.Close)
	return front, got
}

func send(t *testing.T, front *httptest.Server, req *http.Request) int {
	t.Helper()
	u, _ := url.Parse(front.URL)
	req.URL.Scheme, req.URL.Host, req.RequestURI = u.Scheme, u.Host, ""
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

func TestPublicHandler_ForwardsWithoutIdentity(t *testing.T) {
	front, got := publicProxy(t)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/public/blind-tests/invites/tok/answers/img?session=s1&x=1", nil)
	req.Header.Set("X-User-ID", "admin-1")
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("X-Session-ID", "s1")
	req.Header.Set("Authorization", "Bearer user-token")
	req.Header.Set("Cookie", "session_id=abc")
	req.Header.Set("X-Guest-Session", "guest.secret")
	if code := send(t, front, req); code != http.StatusOK {
		t.Fatalf("status %d", code)
	}
	if got.URL.Path != "/api/v1/public/blind-tests/invites/tok/answers/img" {
		t.Errorf("forwarded to %q", got.URL.Path)
	}
	for _, h := range identityHeaders {
		if v := got.Header.Get(h); v != "" {
			t.Errorf("%s reached main-service: %q", h, v)
		}
	}
	if got.Header.Get("X-Guest-Session") != "guest.secret" {
		t.Error("the participant's session token must be forwarded")
	}
	if q := got.URL.Query(); q.Get("session") != "" || q.Get("x") != "1" {
		t.Errorf("query forwarded as %q", got.URL.RawQuery)
	}
}

func TestPublicHandler_RefusesPathsOutsideTheInvitations(t *testing.T) {
	front, got := publicProxy(t)
	for _, p := range []string{
		"/api/v1/public/blind-tests/../../workspaces",
		"/api/v1/public/blind-tests/invites/tok/../../../../workspaces",
		"/api/v1/public/blind-tests/invites%2F..%2F..%2Fworkspaces",
		"/api/v1/public/blind-tests//invites/tok",
	} {
		got.URL = nil
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.URL.Opaque = p // sent exactly as written, not cleaned by the client
		code := send(t, front, req)
		if got.URL != nil {
			t.Errorf("%s reached main-service as %s (status %d)", p, got.URL.Path, code)
		}
	}
}
