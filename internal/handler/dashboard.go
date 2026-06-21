package handler

import (
	"net/http"

	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/render"
	"github.com/tm-ox/go-datastar/internal/stream"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type DashboardHandler struct {
	nav []modules.NavItem
	hub *stream.Hub
}

func NewDashboardHandler(nav []modules.NavItem, hub *stream.Hub) *DashboardHandler {
	return &DashboardHandler{nav: nav, hub: hub}
}

func (h *DashboardHandler) Index(w http.ResponseWriter, r *http.Request) {
	meta := modules.Meta{
		Title:       "Dashboard",
		Description: "Live Wikipedia activity",
	}
	render.Page(w, r, render.View{Nav: h.nav, Path: "/dashboard", Meta: meta, Content: views.DashboardContent()})
}

func (h *DashboardHandler) Stream(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	ch, recent, cancel := h.hub.Subscribe()
	defer cancel()

	for _, ev := range recent { // seed first paint, oldest→newest
		sse.PatchElementTempl(views.FeedRow(ev), datastar.WithSelectorID("feed"), datastar.WithModePrepend())
	}

	for {
		select {
		case ev := <-ch:
			sse.PatchElementTempl(views.FeedRow(ev), datastar.WithSelectorID("feed"), datastar.WithModePrepend())
		case <-r.Context().Done():
			return
		}
	}
}
