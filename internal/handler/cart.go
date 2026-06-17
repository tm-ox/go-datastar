package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/middleware"
	"github.com/tm-ox/go-datastar/internal/store/cart"
	"github.com/tm-ox/go-datastar/internal/store/order"
	"github.com/tm-ox/go-datastar/internal/store/product"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type CartHandler struct {
	nav     []modules.NavItem
	cart    cart.CartStore
	product product.ProductStore
	order   order.OrderStore
}

func NewCartHandler(nav []modules.NavItem, cart cart.CartStore, product product.ProductStore, order order.OrderStore) *CartHandler {
	return &CartHandler{
		nav:     nav,
		cart:    cart,
		product: product,
		order:   order,
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

func (h *CartHandler) DrawerUpdateQty(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		ProductID int `json:"productId"`
		Qty       int `json:"qty"`
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
	if sig.Qty <= 0 {
		if err := h.cart.RemoveItem(c.Value, sig.ProductID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.cart.UpdateQuantity(c.Value, sig.ProductID, sig.Qty); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
	sse.PatchElementTempl(views.CheckoutContent(items, total), datastar.WithSelectorID("checkout-content"), datastar.WithModeOuter())
	sse.MarshalAndPatchSignals(map[string]any{"cartTotal": qty})
}

func (h *CartHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("cart_id")
	if err != nil {
		if r.URL.Query().Has("datastar") {
			sse := datastar.NewSSE(w, r)
			sse.PatchElementTempl(modules.Navbar(h.nav, "/cart"), datastar.WithSelectorID("site-header"), datastar.WithModeInner())
			sse.PatchElementTempl(views.CheckoutContent(nil, 0), datastar.WithSelectorID("main"), datastar.WithModeInner())
			sse.ReplaceURL(url.URL{Path: "/cart"})
			return
		}
		cartTotal := middleware.GetCartTotal(r)
		meta := modules.Meta{Title: "Cart", Description: "Your cart"}
		templ.Handler(views.Checkout(h.nav, "/cart", meta, nil, 0, cartTotal)).ServeHTTP(w, r)
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
	if r.URL.Query().Has("datastar") {
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(modules.Navbar(h.nav, "/cart"), datastar.WithSelectorID("site-header"), datastar.WithModeInner())
		sse.PatchElementTempl(views.CheckoutContent(items, total), datastar.WithSelectorID("main"), datastar.WithModeInner())
		sse.ReplaceURL(url.URL{Path: "/cart"})
		sse.ExecuteScript("window.scrollTo(0,0)")
		return
	}
	cartTotal := middleware.GetCartTotal(r)
	meta := modules.Meta{Title: "Cart", Description: "Your cart"}
	templ.Handler(views.Checkout(h.nav, "/cart", meta, items, total, cartTotal)).ServeHTTP(w, r)
}

func (h *CartHandler) UpdateQty(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		ProductID int `json:"productId"`
		Qty       int `json:"qty"`
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
	if sig.Qty <= 0 {
		if err := h.cart.RemoveItem(c.Value, sig.ProductID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.cart.UpdateQuantity(c.Value, sig.ProductID, sig.Qty); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
	sse.PatchElementTempl(views.CheckoutContent(items, total), datastar.WithSelectorID("main"), datastar.WithModeInner())
	sse.MarshalAndPatchSignals(map[string]any{"cartTotal": qty})
}

func (h *CartHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("cart_id")
	if err != nil {
		http.Error(w, "no cart", http.StatusBadRequest)
		return
	}
	items, err := h.cart.GetItemDetails(c.Value)
	if err != nil || len(items) == 0 {
		http.Error(w, "empty cart", http.StatusBadRequest)
		return
	}
	orderID, err := h.order.Create(c.Value, items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.cart.Clear(c.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/cart/success?order=%d", orderID), http.StatusSeeOther)
}

func (h *CartHandler) Success(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("order")
	cartTotal := middleware.GetCartTotal(r)
	meta := modules.Meta{Title: "Order placed", Description: "Order confirmed"}
	templ.Handler(views.CartSuccess(h.nav, "/cart/success", meta, orderID, cartTotal)).ServeHTTP(w, r)
}
