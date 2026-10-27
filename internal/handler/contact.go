package handler

import (
	"fmt"
	"log"
	"net/http"
	"net/mail"
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
	sender, err := mail.ParseAddress(sig.Email)
	if err != nil {
		sse.PatchElementTempl(views.ContactError("Enter a valid email address."),
			datastar.WithSelectorID("contact-error"), datastar.WithModeOuter())
		return
	}
	sender.Name = strings.TrimSpace(strings.NewReplacer("\r", " ", "\n", " ").Replace(sig.Name))
	body := contactMessage(h.user, h.to, *sender, sig.Message)
	auth := smtp.PlainAuth("", h.user, h.pass, h.host)
	if err := smtp.SendMail(h.host+":"+h.port, auth, h.user, []string{h.to}, body); err != nil {
		log.Printf("contact: send failed: %v", err)
		sse.PatchElementTempl(views.ContactError("Failed to send. Please try again."),
			datastar.WithSelectorID("contact-error"), datastar.WithModeOuter())
		return
	}
	sse.PatchElementTempl(views.ContactSuccess(), datastar.WithSelectorID("contact-form"), datastar.WithModeOuter())
}

// contactMessage builds the email. From must match the authenticated mailbox
// (Spacemail rejects a mismatch with the envelope sender), so the visitor goes
// in Reply-To. mail.Address encodes the name, which blocks header injection.
func contactMessage(from, to string, visitor mail.Address, message string) []byte {
	return []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nReply-To: %s\r\nSubject: Contact from go.tmox.net\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, visitor.String(), message))
}
