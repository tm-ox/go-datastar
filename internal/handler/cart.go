package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/store/cart"
	"github.com/tm-ox/go-datastar/internal/store/product"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type CartHandler struct {
	cart    cart.CartStore
	product product.ProductStore
}

func NewCartHandler(cart cart.CartStore, product product.ProductStore) *CartHandler {
	return &CartHandler{
		cart:    cart,
		product: product,
	}
}

func getOrCreateCartID(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie("cart_id"); err == nil && c.Value != "" {
		return c.Value
	}
	id := uuid.New().String()
	http.SetCookie(w, &http.Cookie{
		Name:     "cart_id",
		Value:    id,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return id
}

func (h *CartHandler) Add(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		ProductID int `json:"productId"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil || sig.ProductID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p, err := h.product.GetByID(sig.ProductID)
	if err != nil || p == nil || p.Stock == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cartID := getOrCreateCartID(w, r)
	if err := h.cart.GetOrCreate(cartID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.cart.AddItem(cartID, sig.ProductID, p.Stock); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	total, err := h.cart.TotalQuantity(cartID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.MarshalAndPatchSignals(map[string]any{"cartTotal": total})
}

func (h *CartHandler) Remove(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		ProductID int `json:"productId"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil || sig.ProductID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c, err := r.Cookie("cart_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.cart.RemoveItem(c.Value, sig.ProductID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	items, err := h.cart.GetItemDetails(c.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	total := 0
	for _, item := range items {
		total += item.Price * item.Quantity
	}
	qty, err := h.cart.TotalQuantity(c.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.CartDrawerItems(items, total), datastar.WithSelectorID("cart-drawer-items"), datastar.WithModeOuter())
	sse.MarshalAndPatchSignals(map[string]any{"cartTotal": qty})
}

func (h *CartHandler) Total(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("cart_id")
	if err != nil {
		datastar.NewSSE(w, r).MarshalAndPatchSignals(map[string]any{"cartTotal": 0})
		return
	}
	total, err := h.cart.TotalQuantity(c.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	datastar.NewSSE(w, r).MarshalAndPatchSignals(map[string]any{"cartTotal": total})
}

func (h *CartHandler) Drawer(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("cart_id")
	if err != nil {
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(views.CartDrawer(nil, 0), datastar.WithSelectorID("cart-drawer"), datastar.WithModeInner())
		return
	}
	items, err := h.cart.GetItemDetails(c.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	total := 0
	for _, item := range items {
		total += item.Price * item.Quantity
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.CartDrawer(items, total), datastar.WithSelectorID("cart-drawer"), datastar.WithModeInner())
}
