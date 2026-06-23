package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/starfederation/datastar-go/datastar"
)

func verifySession(secret []byte, cookie string) bool {
	parts := strings.SplitN(cookie, ".", 2)
	if len(parts) != 2 {
		return false
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[0]))
	got, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	return hmac.Equal(got, mac.Sum(nil))
}

func RequireAuth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("admin_session")
			if err != nil || !verifySession(secret, c.Value) {
				if r.URL.Query().Has("datastar") {
					sse := datastar.NewSSE(w, r)
					sse.MarshalAndPatchSignals(map[string]any{"adminOpen": true, "password": ""})
					sse.ExecuteScript(`document.getElementById('login-error').innerHTML=''`)
				} else {
					http.Redirect(w, r, "/", http.StatusFound)
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
