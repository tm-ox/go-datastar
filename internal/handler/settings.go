package handler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/store/product"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type SettingsHandler struct {
	nav      []modules.NavItem
	sections []modules.NavItem
	store    product.ProductStore
}

func NewSettingsHandler(nav []modules.NavItem, sections []modules.NavItem, store product.ProductStore) *SettingsHandler {
	return &SettingsHandler{nav: nav, sections: sections, store: store}
}

func (h *SettingsHandler) Work(w http.ResponseWriter, r *http.Request) {
	meta := modules.Meta{Title: "Settings — Work", Description: ""}
	templ.Handler(views.SettingsWork(h.nav, h.sections, r.URL.Path, meta)).ServeHTTP(w, r)
}

func (h *SettingsHandler) Shop(w http.ResponseWriter, r *http.Request) {
	meta := modules.Meta{Title: "Settings — Shop", Description: ""}
	products, _, err := h.store.Filter("", false, "", 1, 999)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	categories, err := h.store.UniqueCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	templ.Handler(views.SettingsShop(h.nav, h.sections, r.URL.Path, meta, products, categories)).ServeHTTP(w, r)
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
	products, _, err := h.store.Filter("", false, "", 1, 999)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var updated product.Product
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
		Category string `json:"category"`
		Search   string `json:"search"`
		Page     int    `json:"page"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if sig.Page < 1 {
		sig.Page = 1
	}
	products, _, err := h.store.Filter(sig.Category, false, sig.Search, sig.Page, 999)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.SettingsShopRows(products))
}
