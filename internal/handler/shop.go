package handler

import (
	"net/http"

	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/render"
	"github.com/tm-ox/go-datastar/internal/store"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type productReader interface {
	List(q store.ProductQuery) ([]store.Product, bool, bool, string, string, int, error)
	Filter(q store.ProductQuery) ([]store.Product, bool, bool, string, string, int, error)
	GetByHandle(handle string) (*store.Product, error)
	FilterMeta() (productTypes []string, vendors []string, err error)
}

type ShopHandler struct {
	nav   []modules.NavItem
	store productReader
}

func NewShopHandler(nav []modules.NavItem, store productReader) *ShopHandler {
	return &ShopHandler{nav: nav, store: store}
}

func (h *ShopHandler) Index(w http.ResponseWriter, r *http.Request) {
	q := store.ProductQuery{First: defaultLimit}
	products, hasNext, hasPrev, nextCursor, prevCursor, total, err := h.store.List(q)
	if err != nil {
		http.Error(w, "Products unavailable, try again shortly.", http.StatusBadGateway)
		return
	}
	productTypes, vendors, err := h.store.FilterMeta()
	if err != nil {
		productTypes, vendors = nil, nil
	}
	meta := modules.Meta{Title: "Shop", Description: "Shop"}
	render.Page(w, r, render.View{Nav: h.nav, Path: "/shop", Meta: meta,
		Content: views.ShopContent(products, productTypes, vendors, hasNext, hasPrev, nextCursor, prevCursor, total)})
}

func (h *ShopHandler) Filter(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		ProductType string `json:"productType"`
		Vendor      string `json:"vendor"`
		InStock     bool   `json:"inStock"`
		Search      string `json:"search"`
		NextCursor  string `json:"nextCursor"`
		PrevCursor  string `json:"prevCursor"`
		ShopPage    int    `json:"shopPage"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	q := store.ProductQuery{
		ProductType: sig.ProductType,
		Vendor:      sig.Vendor,
		InStock:     sig.InStock,
		Search:      sig.Search,
		After:       sig.NextCursor,
		Before:      sig.PrevCursor,
		First:       defaultLimit,
		Page:        sig.ShopPage,
	}
	if sig.PrevCursor != "" {
		q.First = 0
		q.Last = defaultLimit
	}
	products, hasNext, hasPrev, nextCursor, prevCursor, total, err := h.store.Filter(q)
	if err != nil {
		http.Error(w, "Products unavailable, try again shortly.", http.StatusBadGateway)
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.ShopGrid(products, hasNext, hasPrev, total))
	sse.MarshalAndPatchSignals(map[string]any{"nextCursor": nextCursor, "prevCursor": prevCursor})
}

func (h *ShopHandler) Detail(w http.ResponseWriter, r *http.Request) {
	handle := r.PathValue("slug")
	p, err := h.store.GetByHandle(handle)
	if err != nil {
		http.Error(w, "Products unavailable, try again shortly.", http.StatusBadGateway)
		return
	}
	if p == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	meta := modules.Meta{Title: p.Name, Description: p.Description}
	render.Page(w, r, render.View{Nav: h.nav, Path: r.URL.Path, Meta: meta, Content: views.ShopDetailContent(p)})
}
