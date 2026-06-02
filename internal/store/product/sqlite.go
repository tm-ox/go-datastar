package product

import "database/sql"

type SQLiteProductStore struct {
	db *sql.DB
}

func NewSQLiteProductStore(db *sql.DB) *SQLiteProductStore {
	return &SQLiteProductStore{db: db}
}

func (s *SQLiteProductStore) List() ([]Product, error) {
	rows, err := s.db.Query("SELECT id, name, description, price, category, slug, created_at FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product

		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category, &p.Slug, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (s *SQLiteProductStore) GetBySlug(slug string) (*Product, error) {
	rows, err := s.db.Query("SELECT id, name, description, price, category, slug, created_at FROM products WHERE slug = ?", slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var p Product
	if rows.Next() {
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category, &p.Slug, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
	}
	return &p, rows.Err()
}
