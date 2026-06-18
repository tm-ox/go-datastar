package handler

import (
	"net/http"

	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/render"
	"github.com/tm-ox/go-datastar/internal/store"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type ShopHandler struct {
	nav   []modules.NavItem
	store *store.ProductStore
}

func NewShopHandler(nav []modules.NavItem, store *store.ProductStore) *ShopHandler {
	return &ShopHandler{nav: nav, store: store}
}

func (h *ShopHandler) Index(w http.ResponseWriter, r *http.Request) {
	products, total, err := h.store.List(1, defaultLimit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	categories, err := h.store.UniqueCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	meta := modules.Meta{Title: "Shop", Description: "Shop"}
	render.Page(w, r, render.View{Nav: h.nav, Path: "/shop", Meta: meta,
		Content: views.ShopContent(products, categories, 1, total, defaultLimit)})
}

func (h *ShopHandler) Filter(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		Category string `json:"category"`
		InStock  bool   `json:"inStock"`
		Page     int    `json:"page"`
		Search   string `json:"search"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	products, total, err := h.store.Filter(store.ProductQuery{
		Category: sig.Category, InStock: sig.InStock,
		Search: sig.Search, Page: sig.Page, Limit: defaultLimit,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.ShopGrid(products, sig.Page, total, defaultLimit))
}

func (h *ShopHandler) Detail(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	p, err := h.store.GetBySlug(slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if p == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	meta := modules.Meta{Title: p.Name, Description: p.Description}
	render.Page(w, r, render.View{Nav: h.nav, Path: r.URL.Path, Meta: meta, Content: views.ShopDetailContent(p)})
}
