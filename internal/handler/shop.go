package handler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/tm-ox/go-datastar/internal/store/product"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type ShopHandler struct {
	nav   []modules.NavItem
	store product.ProductStore
}

func NewShopHandler(nav []modules.NavItem, store product.ProductStore) *ShopHandler {
	return &ShopHandler{nav: nav, store: store}
}

func (h *ShopHandler) Index(w http.ResponseWriter, r *http.Request) {
	meta := modules.Meta{Title: "Shop", Description: "Shop"}
	products, err := h.store.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	templ.Handler(views.Shop(h.nav, "/shop", meta, products)).ServeHTTP(w, r)
}
