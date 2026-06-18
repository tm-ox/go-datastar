package store

import (
	"database/sql"
	"errors"
)

// ErrEmptyCart is returned by Place when there is nothing to order.
var ErrEmptyCart = errors.New("cart is empty")

type Order struct {
	ID        int
	CartID    string
	Status    string
	Total     int
	CreatedAt string
}

type OrderItem struct {
	ID        int
	OrderID   int
	ProductID int
	Price     int
	Quantity  int
	Name      string
}

type OrderStore struct {
	db *sql.DB
}

func NewOrderStore(db *sql.DB) *OrderStore {
	return &OrderStore{db: db}
}

// Place turns a cart into an order: it snapshots the cart's items, persists the
// order and its line items, and empties the cart — all in one transaction, so
// the order and the cleared cart commit together or not at all. It returns
// ErrEmptyCart if there is nothing to order.
func (s *OrderStore) Place(cartID string) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	items, err := itemDetails(tx, cartID)
	if err != nil {
		return 0, err
	}
	if len(items) == 0 {
		return 0, ErrEmptyCart
	}

	res, err := tx.Exec(
		"INSERT INTO orders (cart_id, total) VALUES (?, ?)",
		cartID, subtotal(items),
	)
	if err != nil {
		return 0, err
	}
	orderID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	for _, item := range items {
		if _, err := tx.Exec(
			"INSERT INTO order_items (order_id, product_id, name, price, quantity) VALUES (?, ?, ?, ?, ?)",
			orderID, item.ProductID, item.Name, item.Price, item.Quantity,
		); err != nil {
			return 0, err
		}
	}
	if _, err := tx.Exec("DELETE FROM cart_items WHERE cart_id = ?", cartID); err != nil {
		return 0, err
	}
	return int(orderID), tx.Commit()
}
