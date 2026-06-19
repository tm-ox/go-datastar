package handler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/tm-ox/go-datastar/internal/content"
	"github.com/tm-ox/go-datastar/internal/middleware"
	"github.com/tm-ox/go-datastar/internal/render"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type SiteHandler struct {
	nav         []modules.NavItem
	site        content.SiteContent
	contextHTML string
}

func NewSiteHandler(nav []modules.NavItem, site content.SiteContent) *SiteHandler {
	ctx, err := content.LoadContext()
	if err != nil {
		panic(err)
	}
	return &SiteHandler{
		nav:         nav,
		site:        site,
		contextHTML: string(ctx),
	}
}

func (h *SiteHandler) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		cartTotal := middleware.GetCartTotal(r)
		templ.Handler(views.NotFound(h.nav, r.URL.Path, modules.Meta{Title: "404", Description: "Page not found"}, cartTotal)).ServeHTTP(w, r)
		return
	}
	meta := modules.Meta{
		Title:       h.site.Home.Meta.Title,
		Description: h.site.Home.Meta.Description,
	}
	render.Page(w, r, render.View{Nav: h.nav, Path: "/", Meta: meta, Content: views.IndexContent(h.site.Home)})
}

func (h *SiteHandler) About(w http.ResponseWriter, r *http.Request) {
	meta := modules.Meta{
		Title:       h.site.About.Meta.Title,
		Description: h.site.About.Meta.Description,
	}
	render.Page(w, r, render.View{Nav: h.nav, Path: "/about", Meta: meta, Content: views.AboutContent(h.site.About)})
}

func (h *SiteHandler) Context(w http.ResponseWriter, r *http.Request) {
	meta := modules.Meta{
		Title:       "Context",
		Description: "Site context and conventions",
	}
	render.Page(w, r, render.View{Nav: h.nav, Path: "/context", Meta: meta, Content: views.ContextContent(h.contextHTML)})
}
