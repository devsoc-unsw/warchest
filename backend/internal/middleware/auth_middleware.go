package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	kratos "github.com/ory/kratos-client-go/v26"
)

type contextKey string

const SessionContextKey contextKey = "kratos_session"

type KratosMiddleware struct {
	client    *kratos.APIClient
	publicURL string
}

// NewKratosMiddleware creates a Kratos auth middleware.
// internalURL is used by the backend to call Kratos (e.g. http://mock-kratos:4433).
// publicURL is the browser-facing Kratos URL used for login redirects (e.g. http://localhost:4433).
func NewKratosMiddleware(internalURL, publicURL string) *KratosMiddleware {
	internalURL = strings.TrimRight(internalURL, "/")
	publicURL = strings.TrimRight(publicURL, "/")

	cfg := kratos.NewConfiguration()
	cfg.Servers = []kratos.ServerConfiguration{
		{URL: internalURL},
	}
	return &KratosMiddleware{
		client:    kratos.NewAPIClient(cfg),
		publicURL: publicURL,
	}
}

// RequireAuth validates the Kratos session (via Cookie or X-Session-Token)
// and attaches the session to both the Gin context and the request context.
func (km *KratosMiddleware) RequireAuth() gin.HandlerFunc {
	loginURL := km.publicURL + "/self-service/login/browser"

	return func(c *gin.Context) {
		cookie := strings.Join(c.Request.Header.Values("Cookie"), "; ")
		sessionToken := c.GetHeader("X-Session-Token")
		if cookie == "" && sessionToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":     "unauthorized",
				"login_url": loginURL,
			})
			return
		}

		req := km.client.FrontendAPI.ToSession(c.Request.Context())
		if cookie != "" {
			req = req.Cookie(cookie)
		}
		if sessionToken != "" {
			req = req.XSessionToken(sessionToken)
		}

		session, resp, err := req.Execute()
		if err != nil {
			if resp != nil && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error":     "unauthorized",
					"login_url": loginURL,
				})
				return
			}
			log.Printf("kratos ToSession error: %v", err)
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "authentication service unavailable"})
			return
		}

		if session == nil || !session.GetActive() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":     "session inactive",
				"login_url": loginURL,
			})
			return
		}

		// Attach session to both request context (for services/GetSessionFromContext) and gin.Context.
		ctx := context.WithValue(c.Request.Context(), SessionContextKey, session)
		c.Request = c.Request.WithContext(ctx)
		c.Set(SessionContextKey, session)
		c.Next()
	}
}

// GetSession retrieves the Kratos session from a Gin context.
func GetSession(c *gin.Context) (*kratos.Session, bool) {
	val, exists := c.Get(SessionContextKey)
	if !exists {
		return nil, false
	}
	s, ok := val.(*kratos.Session)
	return s, ok && s != nil
}

// GetSessionFromContext retrieves the Kratos session from a standard context.Context
// (such as c.Request.Context()).
func GetSessionFromContext(ctx context.Context) (*kratos.Session, bool) {
	s, ok := ctx.Value(SessionContextKey).(*kratos.Session)
	return s, ok && s != nil
}
