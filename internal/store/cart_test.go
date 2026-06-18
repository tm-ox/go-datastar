package store

import (
	"testing"

	appdb "github.com/tm-ox/go-datastar/internal/db"
)

func newTestCartStore(t *testing.T) *CartStore {
	t.Helper()
	d, err := appdb.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	d.SetMaxOpenConns(1) // pin to one conn so :memory: is a single DB
	if err := appdb.Migrate(d); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return NewCartStore(d)
}

func TestCartStore_Summary(t *testing.T) {
	s := newTestCartStore(t)
	d := s.db
	if _, err := d.Exec(`INSERT INTO products (name, price, slug, stock) VALUES ('A', 500, 'a', 9), ('B', 250, 'b', 9)`); err != nil {
		t.Fatal(err)
	}
	const cartID = "cart-1"
	if err := s.GetOrCreate(cartID); err != nil {
		t.Fatal(err)
	}
	// 2× product 1 (500c) + 3× product 2 (250c)
	for i := 0; i < 2; i++ {
		if err := s.AddItem(cartID, 1, 9); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 3; i++ {
		if err := s.AddItem(cartID, 2, 9); err != nil {
			t.Fatal(err)
		}
	}

	sum, err := s.Summary(cartID)
	if err != nil {
		t.Fatal(err)
	}
	if len(sum.Items) != 2 {
		t.Errorf("items = %d, want 2 lines", len(sum.Items))
	}
	if sum.Subtotal != 2*500+3*250 {
		t.Errorf("subtotal = %d, want %d", sum.Subtotal, 2*500+3*250)
	}
	if sum.Count != 5 {
		t.Errorf("count = %d, want 5", sum.Count)
	}
}

func TestAddItem_ClampsAtStock(t *testing.T) {
	s := newTestCartStore(t)
	const cartID, productID, maxStock = "cart-1", 1, 2

	if err := s.GetOrCreate(cartID); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ { // add 5 times, stock is 2
		if err := s.AddItem(cartID, productID, maxStock); err != nil {
			t.Fatal(err)
		}
	}

	items, err := s.GetItems(cartID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("want 1 line item, got %d", len(items))
	}
	if items[0].Quantity != maxStock {
		t.Errorf("quantity = %d, want clamped to %d", items[0].Quantity, maxStock)
	}
}
