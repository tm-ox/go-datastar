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
	const cartID = "cart-1"
	if err := s.GetOrCreate(cartID); err != nil {
		t.Fatal(err)
	}
	// 2× product-a (500c) + 3× product-b (250c)
	for i := 0; i < 2; i++ {
		if err := s.AddItem(cartID, "product-a", "Product A", 500, 9); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 3; i++ {
		if err := s.AddItem(cartID, "product-b", "Product B", 250, 9); err != nil {
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
	const cartID, maxStock = "cart-1", 2

	if err := s.GetOrCreate(cartID); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ { // add 5 times, stock cap is 2
		if err := s.AddItem(cartID, "product-a", "Product A", 500, maxStock); err != nil {
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
