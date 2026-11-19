package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestRouter(km *KratosMiddleware, onHit func(c *gin.Context)) *gin.Engine {
	r := gin.New()
	r.GET("/protected", km.RequireAuth(), func(c *gin.Context) {
		if onHit != nil {
			onHit(c)
		}
		c.Status(http.StatusOK)
	})
	return r
}

func TestRequireAuth_NoCredentials(t *testing.T) {
	km := NewKratosMiddleware("http://127.0.0.1:0", "http://localhost:4433/")
	handlerCalled := false
	r := newTestRouter(km, func(c *gin.Context) { handlerCalled = true })

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if handlerCalled {
		t.Fatal("expected protected handler not to be called")
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if body["login_url"] != "http://localhost:4433/self-service/login/browser" {
		t.Fatalf("unexpected login_url: %q", body["login_url"])
	}
}

func TestRequireAuth_ValidCookie(t *testing.T) {
	kratosServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessions/whoami" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Cookie"); got != "ory_kratos_session=valid-cookie" {
			t.Errorf("unexpected Cookie header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "sess-123",
			"active": true,
			"identity": {
				"id": "user-456",
				"schema_id": "default",
				"schema_url": "http://localhost:4433/schemas/default",
				"traits": {"email": "dev@warchest.local"}
			}
		}`))
	}))
	defer kratosServer.Close()

	km := NewKratosMiddleware(kratosServer.URL, "http://localhost:4433")
	var gotGinID, gotCtxID string
	r := newTestRouter(km, func(c *gin.Context) {
		if s, ok := GetSession(c); ok && s.Identity != nil {
			gotGinID = s.Identity.Id
		}
		if s, ok := GetSessionFromContext(c.Request.Context()); ok && s.Identity != nil {
			gotCtxID = s.Identity.Id
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Cookie", "ory_kratos_session=valid-cookie")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	if gotGinID != "user-456" {
		t.Fatalf("GetSession identity mismatch: got %q, want %q", gotGinID, "user-456")
	}
	if gotCtxID != "user-456" {
		t.Fatalf("GetSessionFromContext identity mismatch: got %q, want %q", gotCtxID, "user-456")
	}
}

func TestRequireAuth_ValidSessionToken(t *testing.T) {
	kratosServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Session-Token"); got != "tok-abc" {
			t.Errorf("unexpected X-Session-Token header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "sess-tok",
			"active": true,
			"identity": {
				"id": "user-789",
				"schema_id": "default",
				"schema_url": "http://localhost:4433/schemas/default",
				"traits": {"email": "api@warchest.local"}
			}
		}`))
	}))
	defer kratosServer.Close()

	km := NewKratosMiddleware(kratosServer.URL, "http://localhost:4433")
	var gotID string
	r := newTestRouter(km, func(c *gin.Context) {
		if s, ok := GetSession(c); ok && s.Identity != nil {
			gotID = s.Identity.Id
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-Session-Token", "tok-abc")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotID != "user-789" {
		t.Fatalf("expected identity user-789, got %q", gotID)
	}
}

func TestRequireAuth_InvalidOrInactiveSession(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			name:       "kratos returns 401",
			statusCode: http.StatusUnauthorized,
			body:       `{"error":{"code":401,"message":"No active session was found in this request."}}`,
		},
		{
			name:       "kratos returns inactive session",
			statusCode: http.StatusOK,
			body:       `{"id":"sess-old","active":false}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kratosServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer kratosServer.Close()

			km := NewKratosMiddleware(kratosServer.URL, "http://localhost:4433")
			handlerCalled := false
			r := newTestRouter(km, func(c *gin.Context) { handlerCalled = true })

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Cookie", "ory_kratos_session=bad")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", rec.Code)
			}
			if handlerCalled {
				t.Fatal("expected protected handler not to be called")
			}
		})
	}
}

func TestRequireAuth_KratosUnavailable(t *testing.T) {
	kratosServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer kratosServer.Close()

	km := NewKratosMiddleware(kratosServer.URL, "http://localhost:4433")
	handlerCalled := false
	r := newTestRouter(km, func(c *gin.Context) { handlerCalled = true })

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Cookie", "ory_kratos_session=anything")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}
	if handlerCalled {
		t.Fatal("expected protected handler not to be called")
	}
}
