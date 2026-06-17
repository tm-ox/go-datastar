package handler

import (
	"net/http"
	"net/url"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/middleware"
	"github.com/tm-ox/go-datastar/internal/store/work"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

const defaultWorkLimit = 10

type WorkHandler struct {
	nav   []modules.NavItem
	store work.WorkStore
}

func NewWorkHandler(nav []modules.NavItem, store work.WorkStore) *WorkHandler {
	return &WorkHandler{nav: nav, store: store}
}

func (h *WorkHandler) Index(w http.ResponseWriter, r *http.Request) {
	entries, total, err := h.store.List(1, defaultWorkLimit)
	if err != nil {
		http.Error(w, "failed to load work", http.StatusInternalServerError)
		return
	}
	types, _ := h.store.UniqueTypes()
	clients, _ := h.store.UniqueClients()
	years, _ := h.store.UniqueYears()
	tools, _ := h.store.UniqueTools()

	if r.URL.Query().Has("datastar") {
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(modules.Navbar(h.nav, "/work"), datastar.WithSelectorID("site-header"), datastar.WithModeInner())
		sse.PatchElementTempl(views.WorkContent(entries, total, defaultWorkLimit, types, clients, years, tools), datastar.WithSelectorID("main"), datastar.WithModeInner())
		sse.ReplaceURL(url.URL{Path: "/work"})
		sse.ExecuteScript("window.scrollTo(0,0)")
		return
	}
	meta := modules.Meta{Title: "Work"}
	cartTotal := middleware.GetCartTotal(r)
	templ.Handler(views.Work(h.nav, "/work", meta, entries, total, defaultWorkLimit, types, clients, years, tools, cartTotal)).ServeHTTP(w, r)
}

func (h *WorkHandler) Detail(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	entry, err := h.store.GetBySlug(slug)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if entry == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if r.URL.Query().Has("datastar") {
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(modules.Navbar(h.nav, r.URL.Path), datastar.WithSelectorID("site-header"), datastar.WithModeInner())
		sse.PatchElementTempl(views.WorkDetailContent(entry), datastar.WithSelectorID("main"), datastar.WithModeInner())
		sse.ReplaceURL(url.URL{Path: r.URL.Path})
		sse.ExecuteScript("window.scrollTo(0,0)")
		return
	}
	meta := modules.Meta{Title: entry.Title, Description: entry.Description}
	cartTotal := middleware.GetCartTotal(r)
	templ.Handler(views.WorkDetail(h.nav, r.URL.Path, entry, meta, cartTotal)).ServeHTTP(w, r)
}

func (h *WorkHandler) Filter(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		Type   string `json:"type"`
		Client string `json:"client"`
		Year   string `json:"year"`
		Tool   string `json:"tool"`
		Sort   string `json:"sort"`
		Page   int    `json:"page"`
		Search string `json:"search"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if sig.Page < 1 {
		sig.Page = 1
	}
	entries, total, err := h.store.Filter(sig.Type, sig.Client, sig.Year, sig.Tool, sig.Search, sig.Sort, sig.Page, defaultWorkLimit)
	if err != nil {
		http.Error(w, "filter error", http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.WorkRows(entries, sig.Page, total, defaultWorkLimit))
}
