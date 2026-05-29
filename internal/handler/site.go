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
	templ.Handler(views.Index(h.nav, "/", h.site.Home)).ServeHTTP(w, r)
}

func (h *SiteHandler) About(w http.ResponseWriter, r *http.Request) {
	templ.Handler(views.About(h.nav, "/about", h.site.About)).ServeHTTP(w, r)
}
