package app

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

const sessionCookieName = "sheep_session"

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *App) registerAuthenticationRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/login", a.login)
	mux.HandleFunc("POST /api/logout", a.requireAuthentication(a.logout))
	mux.HandleFunc("GET /api/me", a.requireAuthentication(a.me))
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var input loginRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "Dados de acesso inválidos.")
		return
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	user, authenticated := a.store.authenticate(email, input.Password)
	if !authenticated {
		writeError(w, http.StatusUnauthorized, "E-mail ou senha incorretos.")
		return
	}

	token := newSessionToken()
	a.sessionsMu.Lock()
	a.sessions[token] = user
	a.sessionsMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		MaxAge:   8 * 60 * 60,
	})

	writeJSON(w, http.StatusOK, user)
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		a.sessionsMu.Lock()
		delete(a.sessions, cookie.Value)
		a.sessionsMu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *App) me(w http.ResponseWriter, r *http.Request) {
	user, _ := a.currentUser(r)
	writeJSON(w, http.StatusOK, user)
}

func (a *App) currentUser(r *http.Request) (User, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return User{}, false
	}

	a.sessionsMu.RLock()
	defer a.sessionsMu.RUnlock()

	user, found := a.sessions[cookie.Value]
	return user, found
}

func (a *App) requireAuthentication(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, authenticated := a.currentUser(r); !authenticated {
			writeError(w, http.StatusUnauthorized, "Sessão expirada. Entre novamente.")
			return
		}
		next(w, r)
	}
}

func (a *App) requireModule(module string, next http.HandlerFunc) http.HandlerFunc {
	return a.requireAuthentication(func(w http.ResponseWriter, r *http.Request) {
		user, _ := a.currentUser(r)
		if user.Role != "admin" && !sliceContains(user.Modules, module) {
			writeError(w, http.StatusForbidden, "Seu perfil não possui acesso a este módulo.")
			return
		}
		next(w, r)
	})
}

func newSessionToken() string {
	buffer := make([]byte, 32)
	_, _ = rand.Read(buffer)
	return hex.EncodeToString(buffer)
}
