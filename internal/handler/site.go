package handler

import (
	"net/http"
	"net/url"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
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
	if r.URL.Query().Has("datastar") {
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(modules.Navbar(h.nav, "/"), datastar.WithSelectorID("site-header"), datastar.WithModeInner())
		sse.PatchElementTempl(views.IndexContent(h.site.Home), datastar.WithSelectorID("main"), datastar.WithModeInner())
		sse.ReplaceURL(url.URL{Path: "/"})
		sse.ExecuteScript("window.scrollTo(0,0)")
		return
	}
	templ.Handler(views.Index(h.nav, "/", h.site.Home, meta)).ServeHTTP(w, r)
}

func (h *SiteHandler) About(w http.ResponseWriter, r *http.Request) {
	meta := modules.Meta{
		Title:       h.site.About.Meta.Title,
		Description: h.site.About.Meta.Description,
	}
	if r.URL.Query().Has("datastar") {
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(modules.Navbar(h.nav, "/about"), datastar.WithSelectorID("site-header"), datastar.WithModeInner())
		sse.PatchElementTempl(views.AboutContent(h.site.About), datastar.WithSelectorID("main"), datastar.WithModeInner())
		sse.ReplaceURL(url.URL{Path: "/about"})
		sse.ExecuteScript("window.scrollTo(0,0)")
		return
	}
	templ.Handler(views.About(h.nav, "/about", h.site.About, meta)).ServeHTTP(w, r)
}
