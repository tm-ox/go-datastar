package handler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type ShopHandler struct {
	nav []modules.NavItem
}

func NewShopHandler(nav []modules.NavItem) *ShopHandler {
	return &ShopHandler{nav: nav}
}

func (h *ShopHandler) Index(w http.ResponseWriter, r *http.Request) {
	meta := modules.Meta{Title: "Shop", Description: "Shop"}
	templ.Handler(views.Shop(h.nav, "/shop", meta)).ServeHTTP(w, r)
}
