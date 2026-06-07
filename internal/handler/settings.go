package handler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type SettingsHandler struct {
	nav      []modules.NavItem
	sections []modules.NavItem
}

func NewSettingsHandler(nav []modules.NavItem, sections []modules.NavItem) *SettingsHandler {
	return &SettingsHandler{nav: nav, sections: sections}
}

func (h *SettingsHandler) Work(w http.ResponseWriter, r *http.Request) {
	meta := modules.Meta{Title: "Settings — Work", Description: ""}
	templ.Handler(views.SettingsWork(h.nav, h.sections, r.URL.Path, meta)).ServeHTTP(w, r)
}

func (h *SettingsHandler) Shop(w http.ResponseWriter, r *http.Request) {
	meta := modules.Meta{Title: "Settings — Shop", Description: ""}
	templ.Handler(views.SettingsShop(h.nav, h.sections, r.URL.Path, meta)).ServeHTTP(w, r)
}
