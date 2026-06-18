package store

import (
	"errors"
	"testing"
)

func TestOrderStore_Place(t *testing.T) {
	cart := newTestCartStore(t)
	orders := NewOrderStore(cart.db)
	d := cart.db
	if _, err := d.Exec(`INSERT INTO products (name, price, slug, stock) VALUES ('A', 500, 'a', 9)`); err != nil {
		t.Fatal(err)
	}
	const cartID = "cart-1"
	if err := cart.GetOrCreate(cartID); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := cart.AddItem(cartID, 1, 9); err != nil {
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

	// Order persisted with the correct total.
	var total int
	if err := d.QueryRow("SELECT total FROM orders WHERE id = ?", orderID).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != 3*500 {
		t.Errorf("order total = %d, want %d", total, 3*500)
	}

	// Line items snapshotted.
	var lines int
	if err := d.QueryRow("SELECT COUNT(*) FROM order_items WHERE order_id = ?", orderID).Scan(&lines); err != nil {
		t.Fatal(err)
	}
	if lines != 1 {
		t.Errorf("order_items = %d, want 1", lines)
	}

	// Cart emptied in the same transaction.
	sum, err := cart.Summary(cartID)
	if err != nil {
		t.Fatal(err)
	}
	if sum.Count != 0 {
		t.Errorf("cart count after Place = %d, want 0", sum.Count)
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
