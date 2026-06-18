package store

import "testing"

func newTestProductStore(t *testing.T) *ProductStore {
	t.Helper()
	cart := newTestCartStore(t) // reuses the :memory: + Migrate helper
	return NewProductStore(cart.db)
}

func TestProductStore_Filter(t *testing.T) {
	s := newTestProductStore(t)
	for _, p := range []Product{
		{Name: "Red Mug", Description: "ceramic", Price: 1500, Category: "kitchen", Stock: 3},
		{Name: "Blue Mug", Description: "ceramic", Price: 1800, Category: "kitchen", Stock: 0},
		{Name: "Notebook", Description: "paper", Price: 900, Category: "office", Stock: 5},
	} {
		if _, err := s.Create(p); err != nil {
			t.Fatal(err)
		}
	}

	// Category + in-stock filter via the query object.
	got, total, err := s.Filter(ProductQuery{Category: "kitchen", InStock: true, Page: 1, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(got) != 1 {
		t.Fatalf("kitchen+inStock = %d rows (total %d), want 1", len(got), total)
	}
	if got[0].Name != "Red Mug" {
		t.Errorf("got %q, want Red Mug", got[0].Name)
	}

	// Search across name/description.
	_, total, err = s.Filter(ProductQuery{Search: "ceramic", Page: 1, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("search 'ceramic' total = %d, want 2", total)
	}
}
