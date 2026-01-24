package session

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	SessionCookieName = "templui_session"
	SessionContextKey = "session_token"
)

// GenerateSessionToken creates a new random session token
func GenerateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GetOrCreateSession gets the existing session token or creates a new one
func GetOrCreateSession(c echo.Context) string {
	// Check if session is already in context (from middleware)
	if token := c.Get(SessionContextKey); token != nil {
		return token.(string)
	}

	// Try to get from cookie
	cookie, err := c.Cookie(SessionCookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Create new session
	token := GenerateSessionToken()
	SetSessionCookie(c, token)
	return token
}

// SetSessionCookie sets the session cookie
func SetSessionCookie(c echo.Context, token string) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(365 * 24 * time.Hour), // 1 year
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)
}

// SessionMiddleware ensures every request has a session token
func SessionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := GetOrCreateSession(c)
			c.Set(SessionContextKey, token)
			return next(c)
		}
	}
}

// GetSessionToken gets the session token from the context
func GetSessionToken(c echo.Context) string {
	if token := c.Get(SessionContextKey); token != nil {
		return token.(string)
	}
	return ""
}
