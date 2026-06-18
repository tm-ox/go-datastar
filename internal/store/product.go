package store

import (
	"database/sql"
	"regexp"
	"strings"
)

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

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

type ProductStore struct {
	db *sql.DB
}

func NewProductStore(db *sql.DB) *ProductStore {
	return &ProductStore{db: db}
}

func (s *ProductStore) Create(p Product) (int, error) {
	p.Slug = slugify(p.Name)
	res, err := s.db.Exec(
		"INSERT INTO products (image, name, description, price, category, slug, stock) VALUES (?, ?, ?, ?, ?, ?, ?)",
		p.Image, p.Name, p.Description, p.Price, p.Category, p.Slug, p.Stock,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (s *ProductStore) Update(p Product) error {
	_, err := s.db.Exec(
		"UPDATE products SET image = ?, name = ?, description = ?, price = ?, category = ?, stock = ? WHERE id = ?",
		p.Image, p.Name, p.Description, p.Price, p.Category, p.Stock, p.ID,
	)
	return err
}

func (s *ProductStore) Delete(id int) error {
	_, err := s.db.Exec("DELETE FROM products WHERE id = ?", id)
	return err
}

func (s *ProductStore) GetByID(id int) (*Product, error) {
	var p Product
	err := s.db.QueryRow(
		"SELECT id, image, name, description, price, category, slug, created_at, stock FROM products WHERE id = ?", id,
	).Scan(&p.ID, &p.Image, &p.Name, &p.Description, &p.Price, &p.Category, &p.Slug, &p.CreatedAt, &p.Stock)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (s *ProductStore) List(page, limit int) ([]Product, int, error) {
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

func (s *ProductStore) GetBySlug(slug string) (*Product, error) {
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
	"name-asc":      "name ASC",
	"name-desc":     "name DESC",
	"category-asc":  "category ASC",
	"category-desc": "category DESC",
	"stock-asc":     "stock ASC",
	"stock-desc":    "stock DESC",
}

// ProductQuery describes a filtered, paginated product listing. Zero-valued
// fields are unset: an empty Category matches all categories, a false InStock
// applies no stock filter, and so on.
type ProductQuery struct {
	Category   string
	InStock    bool
	OutOfStock bool
	Sort       string
	Search     string
	Page       int
	Limit      int
}

func (s *ProductStore) Filter(q ProductQuery) ([]Product, int, error) {
	where := []string{}
	args := []any{}

	if q.Category != "" {
		where = append(where, "category = ?")
		args = append(args, q.Category)
	}
	if q.InStock {
		where = append(where, "stock > 0")
	}
	if q.OutOfStock {
		where = append(where, "stock = 0")
	}
	if q.Search != "" {
		where = append(where, "(name LIKE ? OR description LIKE ?)")
		args = append(args, "%"+q.Search+"%", "%"+q.Search+"%")
	}

	baseQuery := "SELECT id, image, name, description, price, category, slug, created_at, stock FROM products"
	countQuery := "SELECT COUNT(*) FROM products"

	if len(where) > 0 {
		clause := " WHERE " + strings.Join(where, " AND ")
		baseQuery += clause
		countQuery += clause
	}

	orderBy := " ORDER BY name ASC"
	if s, ok := sortClauses[q.Sort]; ok {
		orderBy = " ORDER BY " + s
	}
	baseQuery += orderBy

	var total int
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.Limit
	rows, err := s.db.Query(baseQuery+" LIMIT ? OFFSET ?", append(args, q.Limit, offset)...)
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

func (s *ProductStore) UniqueCategories() ([]string, error) {
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

func (s *ProductStore) UpdateStock(id int, stock int) error {
	_, err := s.db.Exec("UPDATE products SET stock = ? WHERE id = ?", stock, id)
	return err
}
