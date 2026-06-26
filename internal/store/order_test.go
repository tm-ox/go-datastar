package store

import (
	"database/sql"
	"errors"
	"testing"
)

func TestOrderStore_Place(t *testing.T) {
	cart := newTestCartStore(t)
	orders := NewOrderStore(cart.db)
	const cartID = "cart-1"
	if err := cart.GetOrCreate(cartID); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := cart.AddItem(cartID, "product-a", "Product A", 500, 9); err != nil {
			t.Fatal(err)
		}
	}

	orderID, err := orders.Place(cartID)
	if err != nil {
		t.Fatalf("Place: %v", err)
	}
	if orderID == 0 {
		t.Fatal("expected a non-zero order id")
	}

	var total int
	if err := cart.db.QueryRow("SELECT total FROM orders WHERE id = ?", orderID).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != 3*500 {
		t.Errorf("order total = %d, want %d", total, 3*500)
	}

	var lines int
	if err := cart.db.QueryRow("SELECT COUNT(*) FROM order_items WHERE order_id = ?", orderID).Scan(&lines); err != nil {
		t.Fatal(err)
	}
	if lines != 1 {
		t.Errorf("order_items = %d, want 1", lines)
	}

	sum, err := cart.Summary(cartID)
	if err != nil {
		t.Fatal(err)
	}
	if sum.Count != 0 {
		t.Errorf("cart count after Place = %d, want 0", sum.Count)
	}
}

func TestOrderStore_GetByID(t *testing.T) {
	cart := newTestCartStore(t)
	orders := NewOrderStore(cart.db)
	const cartID = "cart-getbyid"
	if err := cart.GetOrCreate(cartID); err != nil {
		t.Fatal(err)
	}
	if err := cart.AddItem(cartID, "product-b", "Product B", 1000, 5); err != nil {
		t.Fatal(err)
	}
	orderID, err := orders.Place(cartID)
	if err != nil {
		t.Fatalf("Place: %v", err)
	}

	o, items, err := orders.GetByID(orderID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if o.ID != orderID {
		t.Errorf("order.ID = %d, want %d", o.ID, orderID)
	}
	if o.CartID != cartID {
		t.Errorf("order.CartID = %q, want %q", o.CartID, cartID)
	}
	if o.Total != 1000 {
		t.Errorf("order.Total = %d, want 1000", o.Total)
	}
	if len(items) != 1 {
		t.Errorf("len(items) = %d, want 1", len(items))
	}
}

func TestOrderStore_GetByID_NotFound(t *testing.T) {
	cart := newTestCartStore(t)
	orders := NewOrderStore(cart.db)

	_, _, err := orders.GetByID(999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}
}

func TestOrderStore_Place_EmptyCart(t *testing.T) {
	cart := newTestCartStore(t)
	orders := NewOrderStore(cart.db)

	_, err := orders.Place("nobody")
	if !errors.Is(err, ErrEmptyCart) {
		t.Fatalf("err = %v, want ErrEmptyCart", err)
	}
}
