package handler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/content"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

const workLimit = 10

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
	meta := modules.Meta{
		Title:       "Work",
		Description: "",
	}
	types := content.UniqueTypes(h.entries)
	clients := content.UniqueClients(h.entries)
	years := content.UniqueYears(h.entries)
	tools := content.UniqueTools(h.entries)
	entries := content.FilterWork(h.entries, "", "", "", "", "")
	paged, total := content.PaginateWork(entries, 1, workLimit)
	templ.Handler(views.Work(h.nav, "/work", meta, paged, total, workLimit, types, clients, years, tools)).ServeHTTP(w, r)

}

func (h *WorkHandler) Detail(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	entry, ok := h.bySlug[slug]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	meta := modules.Meta{
		Title:       entry.Title,
		Description: entry.Description,
	}
	templ.Handler(views.WorkDetail(h.nav, r.URL.Path, entry, meta)).ServeHTTP(w, r)
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
	filtered := content.FilterWork(h.entries, sig.Type, sig.Client, sig.Year, sig.Tool, sig.Search)
	filtered = content.SortWork(filtered, sig.Sort)
	paged, total := content.PaginateWork(filtered, sig.Page, workLimit)
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.WorkRows(paged, sig.Page, total, workLimit))
}
