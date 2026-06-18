package middleware

import (
	"context"
	"net/http"

	"github.com/tm-ox/go-datastar/internal/store"
)

type contextKey string

const CartTotalKey contextKey = "cartTotal"

func CartTotal(store *store.CartStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		total := 0
		if c, err := r.Cookie("cart_id"); err == nil {
			total, _ = store.TotalQuantity(c.Value)
		}
		ctx := context.WithValue(r.Context(), CartTotalKey, total)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetCartTotal(r *http.Request) int {
	v, _ := r.Context().Value(CartTotalKey).(int)
	return v
}
