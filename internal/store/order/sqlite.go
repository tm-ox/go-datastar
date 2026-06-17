package order

import (
	"database/sql"

	"github.com/tm-ox/go-datastar/internal/store/cart"
)

type SQLiteOrderStore struct {
	db *sql.DB
}

func NewSQLiteOrderStore(db *sql.DB) *SQLiteOrderStore {
	return &SQLiteOrderStore{db: db}
}

func (s *SQLiteOrderStore) Create(cartID string, items []cart.CartItemDetail) (int, error) {
	total := 0
	for _, item := range items {
		total += item.Price * item.Quantity
	}
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		"INSERT INTO orders (cart_id, total) VALUES (?, ?)",
		cartID, total,
	)
	if err != nil {
		return 0, err
	}
	orderID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	for _, item := range items {
		_, err := tx.Exec(
			"INSERT INTO order_items (order_id, product_id, name, price, quantity) VALUES (?, ?, ?, ?, ?)",
			orderID, item.ProductID, item.Name, item.Price, item.Quantity,
		)
		if err != nil {
			return 0, err
		}
	}
	return int(orderID), tx.Commit()
}
