package handler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

func (h *SiteHandler) Settings(w http.ResponseWriter, r *http.Request) {
	meta := modules.Meta{Title: "Settings", Description: "Settings"}
	templ.Handler(views.Settings(h.nav, "/settings", meta)).ServeHTTP(w, r)
}
