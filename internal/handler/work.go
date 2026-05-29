package handler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/content"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type WorkHandler struct {
	nav     []modules.NavItem
	entries []content.WorkEntry
	bySlug  map[string]content.WorkEntry
}

func NewWorkHandler(nav []modules.NavItem, entries []content.WorkEntry, bySlug map[string]content.WorkEntry) *WorkHandler {
	return &WorkHandler{
		nav:     nav,
		entries: entries,
		bySlug:  bySlug,
	}
}

func (h *WorkHandler) Index(w http.ResponseWriter, r *http.Request) {
	types := content.UniqueTypes(h.entries)
	clients := content.UniqueClients(h.entries)
	years := content.UniqueYears(h.entries)
	tools := content.UniqueTools(h.entries)
	templ.Handler(views.WorkIndex(h.nav, "/work", h.entries, types, clients, years, tools)).ServeHTTP(w, r)
}

func (h *WorkHandler) Detail(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Path[len("/work/"):]
	entry, ok := h.bySlug[slug]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	templ.Handler(views.WorkDetail(h.nav, r.URL.Path, entry)).ServeHTTP(w, r)
}

func (h *WorkHandler) Filter(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		Type   string `json:"type"`
		Client string `json:"client"`
		Year   string `json:"year"`
		Tool   string `json:"tool"`
		Sort   string `json:"sort"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	filtered := content.FilterWork(h.entries, sig.Type, sig.Client, sig.Year, sig.Tool)
	filtered = content.SortWork(filtered, sig.Sort)
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.WorkRows(filtered))
}
