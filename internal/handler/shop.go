package handler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/store/product"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type ShopHandler struct {
	nav   []modules.NavItem
	store product.ProductStore
}

const limit = 9

func NewShopHandler(nav []modules.NavItem, store product.ProductStore) *ShopHandler {
	return &ShopHandler{nav: nav, store: store}
}

func (h *ShopHandler) Index(w http.ResponseWriter, r *http.Request) {
	meta := modules.Meta{Title: "Shop", Description: "Shop"}
	products, total, err := h.store.List(1, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	categories, err := h.store.UniqueCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	templ.Handler(views.Shop(h.nav, "/shop", meta, products, categories, 1, total, limit)).ServeHTTP(w, r)
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
	products, total, err := h.store.Filter(sig.Category, sig.InStock, sig.Search, sig.Page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.ShopGrid(products, sig.Page, total, limit))
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
	templ.Handler(views.ShopDetail(h.nav, r.URL.Path, meta, p)).ServeHTTP(w, r)
}
