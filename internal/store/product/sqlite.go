package product

import "database/sql"

type SQLiteProductStore struct {
	db *sql.DB
}

func NewSQLiteProductStore(db *sql.DB) *SQLiteProductStore {
	return &SQLiteProductStore{db: db}
}

func (s *SQLiteProductStore) List() ([]Product, error) {
	rows, err := s.db.Query("SELECT id, image, name, description, price, category, slug, created_at FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product

		err := rows.Scan(&p.ID, &p.Image, &p.Name, &p.Description, &p.Price, &p.Category, &p.Slug, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (s *SQLiteProductStore) GetBySlug(slug string) (*Product, error) {
	var p Product
	err := s.db.QueryRow("SELECT id, image, name, description, price, category, slug, created_at FROM products WHERE slug = ?", slug).
		Scan(&p.ID, &p.Image, &p.Name, &p.Description, &p.Price, &p.Category, &p.Slug, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *SQLiteProductStore) Filter(category string) ([]Product, error) {
	if category == "" {
		return s.List()
	}
	rows, err := s.db.Query("SELECT id, image, name, description, price, category, slug, created_at FROM products WHERE category = ?", category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.Image, &p.Name, &p.Description, &p.Price, &p.Category, &p.Slug, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (s *SQLiteProductStore) UniqueCategories() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT category FROM products ORDER BY category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		err := rows.Scan(&category)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}
