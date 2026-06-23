package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	nav           []modules.NavItem
	passwordHash  []byte // bcrypt hash, computed at startup
	sessionSecret []byte // HMAC key from SESSION_SECRET env
}

func NewAuthHandler(nav []modules.NavItem, hash, secret []byte) *AuthHandler {
	return &AuthHandler{nav: nav, passwordHash: hash, sessionSecret: secret}
}

func (h *AuthHandler) sign(payload string) string {
	mac := hmac.New(sha256.New, h.sessionSecret)
	mac.Write([]byte(payload))
	return payload + "." + hex.EncodeToString(mac.Sum(nil))
}

func (h *AuthHandler) verify(cookie string) bool {
	parts := strings.SplitN(cookie, ".", 2)
	if len(parts) != 2 {
		return false
	}
	expected := hmac.New(sha256.New, h.sessionSecret)
	expected.Write([]byte(parts[0]))
	got, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	return hmac.Equal(got, expected.Sum(nil))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		Password string `json:"password"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil || sig.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := bcrypt.CompareHashAndPassword(h.passwordHash, []byte(sig.Password))
	if err != nil {
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(views.LoginError("Invalid password"), datastar.WithSelectorID("login-error"), datastar.WithModeInner())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    h.sign("admin"),
		Path:     "/",
		MaxAge:   8 * 60 * 60,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	sse := datastar.NewSSE(w, r)
	sse.ExecuteScript("window.location='/settings/work'")
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
