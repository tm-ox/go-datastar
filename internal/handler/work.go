package handler

import (
	"net/http"

	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/render"
	"github.com/tm-ox/go-datastar/internal/store"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

const defaultWorkLimit = 10

type WorkHandler struct {
	nav   []modules.NavItem
	store *store.WorkStore
}

func NewWorkHandler(nav []modules.NavItem, store *store.WorkStore) *WorkHandler {
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

	meta := modules.Meta{Title: "Work"}
	render.Page(w, r, render.View{Nav: h.nav, Path: "/work", Meta: meta,
		Content: views.WorkContent(entries, total, defaultWorkLimit, types, clients, years, tools)})
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
	meta := modules.Meta{Title: entry.Title, Description: entry.Description}
	render.Page(w, r, render.View{Nav: h.nav, Path: r.URL.Path, Meta: meta, Content: views.WorkDetailContent(entry)})
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
	entries, total, err := h.store.Filter(store.WorkQuery{
		Type: sig.Type, Client: sig.Client, Year: sig.Year, Tool: sig.Tool,
		Search: sig.Search, Sort: sig.Sort, Page: sig.Page, Limit: defaultWorkLimit,
	})
	if err != nil {
		http.Error(w, "filter error", http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.WorkRows(entries, sig.Page, total, defaultWorkLimit))
}
