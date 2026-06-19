// Package render owns the page-shell protocol: given a Surface's content,
// it either renders a full page (BaseLayout around the content) or, when the
// request is a Datastar navigation, patches the shell regions over SSE.
// Handlers name what to show; this module decides how the shell is delivered.
package render

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/middleware"
	"github.com/tm-ox/go-datastar/views/layouts"
	"github.com/tm-ox/go-datastar/views/modules"
)

// View is everything the shell needs to render a Surface.
type View struct {
	Nav     []modules.NavItem
	Path    string
	Meta    modules.Meta
	Content templ.Component
}

// Page renders v as a full page, or — when the request carries the datastar
// query flag — patches the site-header and main regions over SSE, replaces the
// URL, and scrolls to top.
func Page(w http.ResponseWriter, r *http.Request, v View) {
	if r.URL.Query().Has("datastar") {
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(modules.Navbar(v.Nav, v.Path), datastar.WithSelectorID("site-header"), datastar.WithModeInner())
		sse.PatchElementTempl(v.Content, datastar.WithSelectorID("main"), datastar.WithModeInner())
		sse.ReplaceURL(url.URL{Path: v.Path})
		sse.ExecuteScript("window.scrollTo(0,0)")
		sse.ExecuteScript(fmt.Sprintf("document.title = %q", v.Meta.Title))
		return
	}
	cartTotal := middleware.GetCartTotal(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	ctx := templ.WithChildren(r.Context(), v.Content)
	if err := layouts.BaseLayout(v.Nav, v.Path, v.Meta, cartTotal).Render(ctx, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
