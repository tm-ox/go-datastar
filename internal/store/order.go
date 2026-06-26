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
	ID            int
	OrderID       int
	ProductHandle string
	Price         int
	Quantity      int
	Name          string
}

type OrderStore struct {
	db *sql.DB
}

func NewOrderStore(db *sql.DB) *OrderStore {
	return &OrderStore{db: db}
}

func (s *OrderStore) GetByID(id int) (*Order, []OrderItem, error) {
	var o Order
	err := s.db.QueryRow(
		"SELECT id, cart_id, status, total, created_at FROM orders WHERE id = ?", id,
	).Scan(&o.ID, &o.CartID, &o.Status, &o.Total, &o.CreatedAt)
	if err != nil {
		return nil, nil, err
	}

	rows, err := s.db.Query(
		"SELECT id, order_id, product_handle, name, price, quantity FROM order_items WHERE order_id = ?", id,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductHandle, &item.Name, &item.Price, &item.Quantity); err !=
			nil {
			return nil, nil, err
		}
		items = append(items, item)
	}
	return &o, items, rows.Err()
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
			"INSERT INTO order_items (order_id, product_handle, name, price, quantity) VALUES (?, ?, ?, ?, ?)",
			orderID, item.ProductHandle, item.Name, item.Price, item.Quantity,
		); err != nil {
			return 0, err
		}
	}
	if _, err := tx.Exec("DELETE FROM cart_items WHERE cart_id = ?", cartID); err != nil {
		return 0, err
	}
	return int(orderID), tx.Commit()
}
