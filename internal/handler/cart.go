package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/middleware"
	"github.com/tm-ox/go-datastar/internal/render"
	"github.com/tm-ox/go-datastar/internal/store"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"
)

type productStockChecker interface {
	GetByHandle(handle string) (*store.Product, error)
}

type CartHandler struct {
	nav     []modules.NavItem
	cart    *store.CartStore
	product productStockChecker
	order   *store.OrderStore
}

func NewCartHandler(nav []modules.NavItem, cart *store.CartStore, product productStockChecker, order *store.OrderStore) *CartHandler {
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
		ProductHandle string `json:"productHandle"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil || sig.ProductHandle == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p, err := h.product.GetByHandle(sig.ProductHandle)
	if err != nil || p == nil || !p.AvailableForSale {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cartID := getOrCreateCartID(w, r)
	if err := h.cart.GetOrCreate(cartID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.cart.AddItem(cartID, sig.ProductHandle, p.Name, p.Price, p.QuantityAvailable); err != nil {
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

// patchCartFragments emits the fragments a cart mutation can affect: the
// drawer's item list and the checkout section, plus the cart-total signal.
// Patching a target absent from the current page is a no-op, so one response
// serves both the drawer and the checkout page.
func (h *CartHandler) patchCartFragments(w http.ResponseWriter, r *http.Request, sum store.CartSummary) {
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.CartDrawerItems(sum.Items, sum.Subtotal), datastar.WithSelectorID("cart-drawer-items"), datastar.WithModeOuter())
	sse.PatchElementTempl(views.CheckoutContent(sum.Items, sum.Subtotal), datastar.WithSelectorID("checkout-content"), datastar.WithModeOuter())
	sse.MarshalAndPatchSignals(map[string]any{"cartTotal": sum.Count})
}

func (h *CartHandler) Remove(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		ProductHandle string `json:"productHandle"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil || sig.ProductHandle == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c, err := r.Cookie("cart_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.cart.RemoveItem(c.Value, sig.ProductHandle); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sum, err := h.cart.Summary(c.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.patchCartFragments(w, r, sum)
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
	sum, err := h.cart.Summary(c.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sse := datastar.NewSSE(w, r)
	sse.PatchElementTempl(views.CartDrawer(sum.Items, sum.Subtotal), datastar.WithSelectorID("cart-drawer"), datastar.WithModeInner())
}

func (h *CartHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	var sum store.CartSummary
	if c, err := r.Cookie("cart_id"); err == nil {
		sum, err = h.cart.Summary(c.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	meta := modules.Meta{Title: "Cart", Description: "Your cart"}
	render.Page(w, r, render.View{Nav: h.nav, Path: "/cart", Meta: meta, Content: views.CheckoutContent(sum.Items, sum.Subtotal)})
}

// UpdateQty sets a cart line's quantity (removing it when qty drops to zero).
// It serves both the drawer and the checkout page — patchCartFragments updates
// whichever is on screen.
func (h *CartHandler) UpdateQty(w http.ResponseWriter, r *http.Request) {
	var sig struct {
		ProductHandle string `json:"productHandle"`
		Qty           int    `json:"qty"`
	}
	if err := datastar.ReadSignals(r, &sig); err != nil || sig.ProductHandle == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c, err := r.Cookie("cart_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if sig.Qty <= 0 {
		if err := h.cart.RemoveItem(c.Value, sig.ProductHandle); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.cart.UpdateQuantity(c.Value, sig.ProductHandle, sig.Qty); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	sum, err := h.cart.Summary(c.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.patchCartFragments(w, r, sum)
}

func (h *CartHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("cart_id")
	if err != nil {
		http.Error(w, "no cart", http.StatusBadRequest)
		return
	}
	orderID, err := h.order.Place(c.Value)
	if errors.Is(err, store.ErrEmptyCart) {
		http.Error(w, "empty cart", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/cart/success?order=%d", orderID), http.StatusSeeOther)
}

func (h *CartHandler) Success(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("order"))
	if err != nil {
		http.Error(w, "invalid order", http.StatusBadRequest)
		return
	}
	order, items, err := h.order.GetByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c, err := r.Cookie("cart_id")
	if err != nil || c.Value != order.CartID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	cartTotal := middleware.GetCartTotal(r)
	meta := modules.Meta{Title: "Order placed", Description: "Order confirmed"}
	templ.Handler(views.CartSuccess(h.nav, "/cart/success", meta, order, items, cartTotal)).ServeHTTP(w, r)
}
