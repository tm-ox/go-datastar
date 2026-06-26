package handler

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/middleware"
	"github.com/tm-ox/go-datastar/internal/store"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type SettingsHandler struct {
	nav       []modules.NavItem
	sections  []modules.NavItem
	workStore *store.WorkStore
}

func NewSettingsHandler(nav []modules.NavItem, sections []modules.NavItem, workStore *store.WorkStore) *SettingsHandler {
	return &SettingsHandler{nav: nav, sections: sections, workStore: workStore}
}

func (h *SettingsHandler) Work(w http.ResponseWriter, r *http.Request) {
	types, _ := h.workStore.UniqueTypes()
	clients, _ := h.workStore.UniqueClients()
	years, _ := h.workStore.UniqueYears()
	tools, _ := h.workStore.UniqueTools()
	entries, total, err := h.workStore.List(1, defaultLimit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if r.URL.Query().Has("datastar") {
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(modules.Navbar(h.nav, "/settings/work"), datastar.WithSelectorID("site-header"), datastar.WithModeInner())
		sse.PatchElementTempl(views.SettingsWorkContent(h.sections, r.URL.Path, entries, total, defaultLimit, types, clients, years, tools), datastar.WithSelectorID("main"), datastar.WithModeInner())
		sse.ReplaceURL(url.URL{Path: "/settings/work"})
		sse.ExecuteScript("window.scrollTo(0,0)")
		return
	}
	meta := modules.Meta{Title: "Settings — Work", Description: ""}
	cartTotal := middleware.GetCartTotal(r)
	templ.Handler(views.SettingsWork(h.nav, h.sections, r.URL.Path, meta, entries, total, defaultLimit, types, clients, years, tools, cartTotal)).ServeHTTP(w, r)
}

func (h *SettingsHandler) WorkFilter(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		Type   string `json:"type"`
		Client string `json:"client"`
		Year   string `json:"year"`
		Tool   string `json:"tool"`
		Sort   string `json:"sort"`
		Search string `json:"search"`
		Page   int    `json:"page"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if sig.Page < 1 {
		sig.Page = 1
	}
	entries, total, err := h.workStore.Filter(store.WorkQuery{
		Type: sig.Type, Client: sig.Client, Year: sig.Year, Tool: sig.Tool,
		Search: sig.Search, Sort: sig.Sort, Page: sig.Page, Limit: defaultLimit,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.SettingsWorkRows(entries, sig.Page, total, defaultLimit))
}

func (h *SettingsHandler) WorkForm(w http.ResponseWriter, r *http.Request) {
	var entry store.Work
	if idStr := r.URL.Query().Get("id"); idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err == nil {
			found, err := h.workStore.GetByID(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if found != nil {
				entry = *found
			}
		}
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.SettingsWorkForm(entry))
}

func (h *SettingsHandler) WorkCreate(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		SortOrder   int    `json:"sortOrder"`
		Title       string `json:"title"`
		WorkType    string `json:"type"`
		Client      string `json:"client"`
		Year        int    `json:"year"`
		Tools       string `json:"tools"`
		Description string `json:"description"`
		Website     string `json:"website"`
		Link        string `json:"link"`
		CoverURL    string `json:"coverURL"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	entry := store.Work{
		SortOrder:   sig.SortOrder,
		Title:       sig.Title,
		WorkType:    sig.WorkType,
		Client:      sig.Client,
		Year:        sig.Year,
		Tools:       sig.Tools,
		Description: sig.Description,
		Website:     sig.Website,
		Link:        sig.Link,
		CoverURL:    sig.CoverURL,
	}
	id, err := h.workStore.Create(entry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	created, err := h.workStore.GetByID(id)
	if err != nil || created == nil {
		http.Error(w, "entry not found after create", http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	entries, total, err := h.workStore.List(1, defaultLimit)
	if err == nil {
		sse.PatchElementTempl(views.SettingsWorkRows(entries, 1, total, defaultLimit))
	}
	sse.MarshalAndPatchSignals(map[string]any{"modalOpen": false})
}

func (h *SettingsHandler) WorkUpdate(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		ID          int    `json:"id"`
		SortOrder   int    `json:"sortOrder"`
		Title       string `json:"title"`
		WorkType    string `json:"type"`
		Client      string `json:"client"`
		Year        int    `json:"year"`
		Tools       string `json:"tools"`
		Description string `json:"description"`
		Website     string `json:"website"`
		Link        string `json:"link"`
		CoverURL    string `json:"coverURL"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	entry := store.Work{
		ID:          sig.ID,
		SortOrder:   sig.SortOrder,
		Title:       sig.Title,
		WorkType:    sig.WorkType,
		Client:      sig.Client,
		Year:        sig.Year,
		Tools:       sig.Tools,
		Description: sig.Description,
		Website:     sig.Website,
		Link:        sig.Link,
		CoverURL:    sig.CoverURL,
	}
	if err := h.workStore.Update(entry); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	updated, err := h.workStore.GetByID(sig.ID)
	if err != nil || updated == nil {
		http.Error(w, "entry not found after update", http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.SettingsWorkRow(*updated))
	sse.MarshalAndPatchSignals(map[string]any{"modalOpen": false})
}

func (h *SettingsHandler) WorkDelete(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		ID int `json:"id"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.workStore.Delete(sig.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElements("", datastar.WithSelectorID("work-"+strconv.Itoa(sig.ID)), datastar.WithModeRemove())
	entries, total, err := h.workStore.List(1, defaultLimit)
	if err == nil {
		sse.PatchElementTempl(views.SettingsWorkRows(entries, 1, total, defaultLimit))
	}
	sse.MarshalAndPatchSignals(map[string]any{"modalOpen": false})
}

func (h *SettingsHandler) Shop(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("datastar") {
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(modules.Navbar(h.nav, "/settings/shop"), datastar.WithSelectorID("site-header"), datastar.WithModeInner())
		sse.PatchElementTempl(views.SettingsShop(h.nav, h.sections, r.URL.Path, modules.Meta{Title: "Settings — Shop"}, middleware.GetCartTotal(r)), datastar.WithSelectorID("main"), datastar.WithModeInner())
		sse.ReplaceURL(url.URL{Path: "/settings/shop"})
		sse.ExecuteScript("window.scrollTo(0,0)")
		return
	}
	meta := modules.Meta{Title: "Settings — Shop", Description: ""}
	cartTotal := middleware.GetCartTotal(r)
	templ.Handler(views.SettingsShop(h.nav, h.sections, r.URL.Path, meta, cartTotal)).ServeHTTP(w, r)
}
