package product

import (
	"database/sql"
	"strings"
)

type SQLiteProductStore struct {
	db *sql.DB
}

func NewSQLiteProductStore(db *sql.DB) *SQLiteProductStore {
	return &SQLiteProductStore{db: db}
}

func (s *SQLiteProductStore) List(page, limit int) ([]Product, int, error) {
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM products").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	rows, err := s.db.Query("SELECT id, image, name, description, price, category, slug, created_at, stock FROM products LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.Image, &p.Name, &p.Description, &p.Price, &p.Category, &p.Slug, &p.CreatedAt, &p.Stock)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	return products, total, rows.Err()
}

func (s *SQLiteProductStore) GetBySlug(slug string) (*Product, error) {
	var p Product
	err := s.db.QueryRow("SELECT id, image, name, description, price, category, slug, created_at, stock FROM products WHERE slug = ?", slug).
		Scan(&p.ID, &p.Image, &p.Name, &p.Description, &p.Price, &p.Category, &p.Slug, &p.CreatedAt, &p.Stock)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

var sortClauses = map[string]string{
	"name-asc":     "name ASC",
	"name-desc":    "name DESC",
	"category-asc": "category ASC",
	"category-desc": "category DESC",
	"stock-asc":    "stock ASC",
	"stock-desc":   "stock DESC",
}

func (s *SQLiteProductStore) Filter(category string, inStock bool, outOfStock bool, sort string, search string, page, limit int) ([]Product, int, error) {
	where := []string{}
	args := []any{}

	if category != "" {
		where = append(where, "category = ?")
		args = append(args, category)
	}
	if inStock {
		where = append(where, "stock > 0")
	}
	if outOfStock {
		where = append(where, "stock = 0")
	}
	if search != "" {
		where = append(where, "(name LIKE ? OR description LIKE ?)")
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	baseQuery := "SELECT id, image, name, description, price, category, slug, created_at, stock FROM products"
	countQuery := "SELECT COUNT(*) FROM products"

	if len(where) > 0 {
		clause := " WHERE " + strings.Join(where, " AND ")
		baseQuery += clause
		countQuery += clause
	}

	orderBy := " ORDER BY name ASC"
	if s, ok := sortClauses[sort]; ok {
		orderBy = " ORDER BY " + s
	}
	baseQuery += orderBy

	var total int
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	rows, err := s.db.Query(baseQuery+" LIMIT ? OFFSET ?", append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Image, &p.Name, &p.Description, &p.Price, &p.Category, &p.Slug, &p.CreatedAt, &p.Stock); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	return products, total, rows.Err()
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
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

func (s *SQLiteProductStore) UpdateStock(id int, stock int) error {
	_, err := s.db.Exec("UPDATE products SET stock = ? WHERE id = ?", stock, id)
	return err
}
