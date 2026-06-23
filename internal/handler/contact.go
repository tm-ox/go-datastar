package handler

import (
	"fmt"
	"net/http"
	"net/smtp"
	"strings"

	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type ContactHandler struct {
	nav  []modules.NavItem
	host string
	port string
	user string
	pass string
	to   string
}

func NewContactHandler(nav []modules.NavItem, host, port, user, pass string) *ContactHandler {
	return &ContactHandler{nav: nav, host: host, port: port, user: user, pass: pass, to: "hello@tmox.net"}
}

func (h *ContactHandler) Send(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Message  string `json:"message"`
		Honeypot string `json:"honeypot"`
	}
	readErr := datastar.ReadSignals(r, &sig)
	sse := datastar.NewSSE(w, r)
	if readErr != nil {
		sse.PatchElementTempl(views.ContactError("Something went wrong. Try again."),
			datastar.WithSelectorID("contact-error"), datastar.WithModeOuter())
		return
	}
	if sig.Honeypot != "" {
		sse.PatchElementTempl(views.ContactSuccess(),
			datastar.WithSelectorID("contact-form"), datastar.WithModeOuter())
		return
	}
	if strings.TrimSpace(sig.Name) == "" || strings.TrimSpace(sig.Email) == "" || strings.TrimSpace(sig.Message) == "" {
		sse.PatchElementTempl(views.ContactError("All fields are required."),
			datastar.WithSelectorID("contact-error"), datastar.WithModeOuter())
		return
	}
	body := fmt.Sprintf("From: %s <%s>\r\nSubject: Contact from go.tmox.net\r\n\r\n%s", sig.Name, sig.Email, sig.Message)
	auth := smtp.PlainAuth("", h.user, h.pass, h.host)
	err := smtp.SendMail(h.host+":"+h.port, auth, h.user, []string{h.to}, []byte(body))
	if err != nil {
		sse.PatchElementTempl(views.ContactError("Failed to send. Please try again."),
			datastar.WithSelectorID("contact-error"), datastar.WithModeOuter())
		return
	}
	sse.PatchElementTempl(views.ContactSuccess(), datastar.WithSelectorID("contact-form"), datastar.WithModeOuter())
}
