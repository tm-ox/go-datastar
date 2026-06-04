package product

import "database/sql"

type Product struct {
	ID          int
	Image       sql.NullString
	Name        string
	Description string
	Price       int
	Category    string
	Slug        string
	CreatedAt   string
	Stock       int
}

type ProductStore interface {
	List(page, limit int) ([]Product, int, error)
	GetBySlug(slug string) (*Product, error)
	Filter(category string, inStock bool, page, limit int) ([]Product, int, error)
	UniqueCategories() ([]string, error)
}
