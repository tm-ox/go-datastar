package handler

import (
	"net/http"
	"time"

	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/render"
	"github.com/tm-ox/go-datastar/internal/stream"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type DashboardHandler struct {
	nav []modules.NavItem
	hub *stream.Hub
	agg *stream.Aggregator
}

func NewDashboardHandler(nav []modules.NavItem, hub *stream.Hub, agg *stream.Aggregator) *DashboardHandler {
	return &DashboardHandler{nav: nav, hub: hub, agg: agg}
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

	stats := h.agg.Snapshot() // snapshot on connect
	lastTotal := stats.TotalEdits
	sse.PatchElementTempl(views.StatsTiles(stats, 0), datastar.WithSelectorID("stats"), datastar.WithModeInner())
	sse.PatchElementTempl(views.Leaderboard(stats.TopWikis), datastar.WithSelectorID("leaderboard"), datastar.WithModeInner())
	sse.PatchElementTempl(views.Sparkline(stats.Sparkline), datastar.WithSelectorID("sparkline"), datastar.WithModeInner())

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case ev := <-ch:
			sse.PatchElementTempl(views.FeedRow(ev), datastar.WithSelectorID("feed"), datastar.WithModePrepend())
		case <-ticker.C:
			stats = h.agg.Snapshot()
			rate := stats.TotalEdits - lastTotal // edits since last tick = per-second rate
			lastTotal = stats.TotalEdits
			sse.PatchElementTempl(views.StatsTiles(stats, rate), datastar.WithSelectorID("stats"), datastar.WithModeInner())
			sse.PatchElementTempl(views.Leaderboard(stats.TopWikis), datastar.WithSelectorID("leaderboard"), datastar.WithModeInner())
			sse.PatchElementTempl(views.Sparkline(stats.Sparkline), datastar.WithSelectorID("sparkline"), datastar.WithModeInner())
		case <-r.Context().Done():
			return
		}
	}
}
