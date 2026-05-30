package handler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/tm-ox/go-datastar/internal/content"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type SiteHandler struct {
	nav  []modules.NavItem
	site content.SiteContent
}

func NewSiteHandler(nav []modules.NavItem, site content.SiteContent) *SiteHandler {
	return &SiteHandler{
		nav:  nav,
		site: site,
	}
}

func (h *SiteHandler) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		templ.Handler(views.NotFound(h.nav, r.URL.Path, modules.Meta{Title: "404", Description: "Page not found"})).ServeHTTP(w, r)
		return
	}
	meta := modules.Meta{
		Title:       h.site.Home.Meta.Title,
		Description: h.site.Home.Meta.Description,
	}
	templ.Handler(views.Index(h.nav, "/", h.site.Home, meta)).ServeHTTP(w, r)
}

func (h *SiteHandler) About(w http.ResponseWriter, r *http.Request) {
	meta := modules.Meta{
		Title:       h.site.About.Meta.Title,
		Description: h.site.About.Meta.Description,
	}
	templ.Handler(views.About(h.nav, "/about", h.site.About, meta)).ServeHTTP(w, r)
}
