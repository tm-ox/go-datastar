package handler

import (
	"net/mail"
	"strings"
	"testing"
)

func TestContactMessage(t *testing.T) {
	msg := string(contactMessage("tim@tmox.net", "hello@tmox.net",
		mail.Address{Name: "Test", Address: "test@test.com"}, "hi"))
	for _, want := range []string{
		"From: tim@tmox.net\r\n",
		"To: hello@tmox.net\r\n",
		"Reply-To: \"Test\" <test@test.com>\r\n",
		"\r\n\r\nhi",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("message missing %q:\n%s", want, msg)
		}
	}
}

func TestContactMessageNoHeaderInjection(t *testing.T) {
	name := cleanName("Eve\r\nBcc: x@y.com")
	msg := string(contactMessage("tim@tmox.net", "hello@tmox.net",
		mail.Address{Name: name, Address: "e@e.com"}, "hi"))
	if strings.Contains(msg, "\r\nBcc:") {
		t.Errorf("injected header present:\n%s", msg)
	}
}

func TestContactMessageCRLFBody(t *testing.T) {
	msg := string(contactMessage("tim@tmox.net", "hello@tmox.net",
		mail.Address{Address: "e@e.com"}, "a\nb\r\nc"))
	if !strings.HasSuffix(msg, "\r\n\r\na\r\nb\r\nc") {
		t.Errorf("body not CRLF-normalised: %q", msg)
	}
}
