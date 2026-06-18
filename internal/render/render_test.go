package render

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/tm-ox/go-datastar/views/modules"
)

func testView() View {
	return View{
		Nav:     nil,
		Path:    "/work",
		Meta:    modules.Meta{Title: "Work"},
		Content: templ.Raw(`<p id="probe">hello</p>`),
	}
}

func TestPage_FullRenderWhenNotDatastar(t *testing.T) {
	r := httptest.NewRequest("GET", "/work", nil)
	w := httptest.NewRecorder()

	Page(w, r, testView())

	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
	body := w.Body.String()
	if !strings.Contains(strings.ToLower(body), "<!doctype html>") {
		t.Error("full render should emit a full document (BaseLayout)")
	}
	if !strings.Contains(body, `id="probe"`) {
		t.Error("full render should contain the Content")
	}
}

func TestPage_SSEPatchesShellWhenDatastar(t *testing.T) {
	r := httptest.NewRequest("GET", "/work?datastar=true", nil)
	w := httptest.NewRecorder()

	Page(w, r, testView())

	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Errorf("content-type = %q, want text/event-stream", ct)
	}
	body := w.Body.String()
	if strings.Contains(strings.ToLower(body), "<!doctype html>") {
		t.Error("datastar nav must not emit a full document")
	}
	for _, want := range []string{"site-header", "main", `id="probe"`} {
		if !strings.Contains(body, want) {
			t.Errorf("SSE stream missing %q", want)
		}
	}
}
