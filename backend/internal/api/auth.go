package api

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"onestarlab/backend/internal/auth"
	"onestarlab/backend/internal/store"
)

const (
	sessionCookieName  = "onestarlab_session"
	maxLoginFailures   = 5
	loginBlockDuration = time.Minute
)

type loginEntry struct {
	failures     int
	lastFailure  time.Time
	blockedUntil time.Time
}

type loginLimiter struct {
	mu      sync.Mutex
	entries map[string]*loginEntry
}

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{entries: make(map[string]*loginEntry)}
}

// allowed 返回是否允许尝试登录，以及被锁定时的剩余时间。
func (l *loginLimiter) allowed(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.entries[key]
	if !ok {
		return true, 0
	}
	if time.Now().Before(entry.blockedUntil) {
		return false, time.Until(entry.blockedUntil)
	}
	return true, 0
}

func (l *loginLimiter) failed(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.entries) > 2048 {
		l.entries = make(map[string]*loginEntry)
	}
	entry, ok := l.entries[key]
	if !ok {
		entry = &loginEntry{}
		l.entries[key] = entry
	}
	if time.Since(entry.lastFailure) > loginBlockDuration {
		entry.failures = 0
	}
	entry.failures++
	entry.lastFailure = time.Now()
	if entry.failures >= maxLoginFailures {
		entry.failures = 0
		entry.blockedUntil = time.Now().Add(loginBlockDuration)
	}
}

func (l *loginLimiter) reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		first := strings.TrimSpace(strings.Split(forwarded, ",")[0])
		if first != "" {
			return first
		}
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

func requestIsSecure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https")
}

func (s *Server) setSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   requestIsSecure(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(auth.SessionTTL.Seconds()),
	})
}

func (s *Server) clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   requestIsSecure(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func (s *Server) hasValidSession(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	secret, err := s.store.SessionSecret()
	if err != nil {
		return false
	}
	return auth.ValidateSession(secret, cookie.Value)
}

func isPublicAPIPath(path string) bool {
	switch path {
	case "/api/v1/health",
		"/api/v1/auth/login",
		"/api/v1/auth/logout",
		"/api/v1/auth/status":
		return true
	default:
		return false
	}
}

func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": s.hasValidSession(r),
		"usingDefault":  s.store.PasscodeIsDefault(),
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	key := clientIP(r)
	if allowed, wait := s.loginLimiter.allowed(key); !allowed {
		writeError(w, http.StatusTooManyRequests, fmt.Sprintf("尝试次数过多，请 %d 秒后再试", int(wait.Seconds())+1))
		return
	}

	var body struct {
		Passcode string `json:"passcode"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	passcode := strings.TrimSpace(body.Passcode)
	if passcode == "" {
		writeError(w, http.StatusBadRequest, "请输入口令")
		return
	}

	ok, err := s.store.VerifyPasscode(passcode)
	if err != nil {
		fail(w, err)
		return
	}
	if !ok {
		s.loginLimiter.failed(key)
		writeError(w, http.StatusUnauthorized, "口令不正确")
		return
	}
	s.loginLimiter.reset(key)

	secret, err := s.store.SessionSecret()
	if err != nil {
		fail(w, err)
		return
	}
	token, err := auth.SignSession(secret, auth.SessionTTL)
	if err != nil {
		fail(w, err)
		return
	}

	s.setSessionCookie(w, r, token)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "usingDefault": s.store.PasscodeIsDefault()})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.clearSessionCookie(w, r)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleChangePasscode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Current string `json:"current"`
		Next    string `json:"next"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}

	err := s.store.ChangePasscode(strings.TrimSpace(body.Current), strings.TrimSpace(body.Next))
	if err != nil {
		if errors.Is(err, store.ErrInvalidPasscode) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if message := err.Error(); strings.Contains(message, "口令") {
			writeError(w, http.StatusBadRequest, message)
			return
		}
		fail(w, err)
		return
	}

	secret, err := s.store.SessionSecret()
	if err != nil {
		fail(w, err)
		return
	}
	token, err := auth.SignSession(secret, auth.SessionTTL)
	if err != nil {
		fail(w, err)
		return
	}
	s.setSessionCookie(w, r, token)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
