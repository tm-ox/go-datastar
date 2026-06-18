package store

import (
	"testing"

	appdb "github.com/tm-ox/go-datastar/internal/db"
)

func newTestWorkStore(t *testing.T) *WorkStore {
	t.Helper()
	d, err := appdb.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	d.SetMaxOpenConns(1)
	if err := appdb.Migrate(d); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return NewWorkStore(d)
}

func TestWorkStore_CreateRoundTrips(t *testing.T) {
	s := newTestWorkStore(t)

	id, err := s.Create(Work{
		Title:       "Brand Refresh",
		SortOrder:   1,
		WorkType:    "branding",
		Client:      "Acme",
		Year:        2026,
		Tools:       "Figma,Penpot",
		Description: "A refresh.",
		Website:     "https://acme.test",
		Link:        "https://acme.test/work",
		CoverURL:    "https://acme.test/cover.jpg",
	})
	if err != nil {
		t.Fatalf("Create: %v", err) // would fail red with the 10-placeholder bug
	}

	got, err := s.GetByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("work not found after create")
	}
	if got.Title != "Brand Refresh" || got.Slug != "brand-refresh" {
		t.Errorf("title/slug = %q/%q, want Brand Refresh/brand-refresh", got.Title, got.Slug)
	}
	if got.Client != "Acme" || got.Year != 2026 {
		t.Errorf("client/year = %q/%d, want Acme/2026", got.Client, got.Year)
	}
}
