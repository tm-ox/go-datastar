package handler

import (
	"database/sql"
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
	store     *store.ProductStore
	workStore *store.WorkStore
}

func NewSettingsHandler(nav []modules.NavItem, sections []modules.NavItem, store *store.ProductStore, workStore *store.WorkStore) *SettingsHandler {
	return &SettingsHandler{nav: nav, sections: sections, store: store, workStore: workStore}
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
	products, total, err := h.store.Filter(store.ProductQuery{Page: 1, Limit: defaultLimit})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	categories, err := h.store.UniqueCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if r.URL.Query().Has("datastar") {
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(modules.Navbar(h.nav, "/settings/shop"), datastar.WithSelectorID("site-header"), datastar.WithModeInner())
		sse.PatchElementTempl(views.SettingsShopContent(h.sections, r.URL.Path, products, categories, total, defaultLimit), datastar.WithSelectorID("main"), datastar.WithModeInner())
		sse.ReplaceURL(url.URL{Path: "/settings/shop"})
		sse.ExecuteScript("window.scrollTo(0,0)")
		return
	}
	meta := modules.Meta{Title: "Settings — Shop", Description: ""}
	cartTotal := middleware.GetCartTotal(r)
	templ.Handler(views.SettingsShop(h.nav, h.sections, r.URL.Path, meta, products, categories, total, defaultLimit, cartTotal)).ServeHTTP(w, r)
}

func (h *SettingsHandler) ShopStock(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		ProductId    int `json:"productId"`
		StockValue   int `json:"stockValue"`
		CurrentStock int `json:"currentStock"`
		Delta        int `json:"delta"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	newStock := sig.StockValue
	if sig.Delta != 0 {
		newStock = sig.CurrentStock + sig.Delta
	}
	if newStock < 0 {
		newStock = 0
	}
	if err := h.store.UpdateStock(sig.ProductId, newStock); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	products, _, err := h.store.Filter(store.ProductQuery{Page: 1, Limit: defaultLimit})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var updated store.Product
	for _, p := range products {
		if p.ID == sig.ProductId {
			updated = p
			break
		}
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.SettingsShopRow(updated))
}

func (h *SettingsHandler) ShopFilter(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		Category   string `json:"category"`
		Search     string `json:"search"`
		Page       int    `json:"page"`
		OutOfStock bool   `json:"outOfStock"`
		Sort       string `json:"sort"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if sig.Page < 1 {
		sig.Page = 1
	}
	products, total, err := h.store.Filter(store.ProductQuery{
		Category: sig.Category, OutOfStock: sig.OutOfStock, Sort: sig.Sort,
		Search: sig.Search, Page: sig.Page, Limit: defaultLimit,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.SettingsShopRows(products, sig.Page, total, defaultLimit))
}

func (h *SettingsHandler) ShopProductForm(w http.ResponseWriter, r *http.Request) {
	var p store.Product
	if idStr := r.URL.Query().Get("id"); idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err == nil {
			found, err := h.store.GetByID(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if found != nil {
				p = *found
			}
		}
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.SettingsShopProductForm(p))
}

func (h *SettingsHandler) ShopProductCreate(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Category    string  `json:"category"`
		Image       string  `json:"image"`
		Stock       int     `json:"stock"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p := store.Product{
		Name:        sig.Name,
		Description: sig.Description,
		Price:       int(sig.Price * 100),
		Category:    sig.Category,
		Stock:       sig.Stock,
	}
	if sig.Image != "" {
		p.Image = sql.NullString{String: sig.Image, Valid: true}
	}
	id, err := h.store.Create(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	created, err := h.store.GetByID(id)
	if err != nil || created == nil {
		http.Error(w, "product not found after create", http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	categories, err := h.store.UniqueCategories()
	if err == nil {
		sse.PatchElementTempl(views.SettingsShopCategories(categories))
	}
	sse.MarshalAndPatchSignals(map[string]any{"modalOpen": false, "category": ""})
	products, total, err := h.store.Filter(store.ProductQuery{Page: 1, Limit: defaultLimit})
	if err == nil {
		sse.PatchElementTempl(views.SettingsShopRows(products, 1, total, defaultLimit))
	}
	sse.MarshalAndPatchSignals(map[string]any{"modalOpen": false, "category": "", "search": ""})
}

func (h *SettingsHandler) ShopProductUpdate(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		ID          int     `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Category    string  `json:"category"`
		Image       string  `json:"image"`
		Stock       int     `json:"stock"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p := store.Product{
		ID:          sig.ID,
		Name:        sig.Name,
		Description: sig.Description,
		Price:       int(sig.Price * 100),
		Category:    sig.Category,
		Stock:       sig.Stock,
	}
	if sig.Image != "" {
		p.Image = sql.NullString{String: sig.Image, Valid: true}
	}
	if err := h.store.Update(p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	updated, err := h.store.GetByID(sig.ID)
	if err != nil || updated == nil {
		http.Error(w, "product not found after update", http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.SettingsShopRow(*updated))
	sse.MarshalAndPatchSignals(map[string]any{"modalOpen": false})
}

func (h *SettingsHandler) ShopProductDelete(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		ID int `json:"id"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.store.Delete(sig.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElements("", datastar.WithSelectorID("product-"+strconv.Itoa(sig.ID)), datastar.WithModeRemove())

	categories, err := h.store.UniqueCategories()
	if err == nil {
		sse.PatchElementTempl(views.SettingsShopCategories(categories))
	}
	sse.MarshalAndPatchSignals(map[string]any{"category": ""})
	products, total, err := h.store.Filter(store.ProductQuery{Page: 1, Limit: defaultLimit})
	if err == nil {
		sse.PatchElementTempl(views.SettingsShopRows(products, 1, total, defaultLimit))
	}
	sse.MarshalAndPatchSignals(map[string]any{"category": "", "search": ""})
}
