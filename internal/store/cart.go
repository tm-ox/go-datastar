package store

import "database/sql"

type CartItem struct {
	ID        int
	CartID    string
	ProductID int
	Quantity  int
}

type CartItemDetail struct {
	ID        int
	ProductID int
	Name      string
	Price     int
	Quantity  int
}

// CartSummary is the computed view of a Cart: its line items plus the money
// subtotal (cents) and item count derived from them. Cart owns this arithmetic
// so callers read it rather than re-summing line items themselves.
type CartSummary struct {
	Items    []CartItemDetail
	Subtotal int
	Count    int
}

// dbtx is the read surface shared by *sql.DB and *sql.Tx, so a query can run
// either on the pool or inside a transaction.
type dbtx interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// itemDetails loads a cart's line items joined to their products. Shared by
// CartStore.GetItemDetails (pool) and OrderStore.Place (in-transaction).
func itemDetails(q dbtx, cartID string) ([]CartItemDetail, error) {
	rows, err := q.Query(`
                SELECT ci.id, ci.product_id, p.name, p.price, ci.quantity
                FROM cart_items ci
                JOIN products p ON p.id = ci.product_id
                WHERE ci.cart_id = ?`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CartItemDetail
	for rows.Next() {
		var item CartItemDetail
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Name, &item.Price, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// subtotal is the single home for cart money arithmetic, in cents.
func subtotal(items []CartItemDetail) int {
	total := 0
	for _, it := range items {
		total += it.Price * it.Quantity
	}
	return total
}

type CartStore struct {
	db *sql.DB
}

func NewCartStore(db *sql.DB) *CartStore {
	return &CartStore{db: db}
}

// Summary loads the cart's line items and computes its subtotal and count in
// one place. Count is derived from the joined items, so an item whose product
// no longer exists neither shows nor counts.
func (s *CartStore) Summary(cartID string) (CartSummary, error) {
	items, err := s.GetItemDetails(cartID)
	if err != nil {
		return CartSummary{}, err
	}
	sum := CartSummary{Items: items, Subtotal: subtotal(items)}
	for _, it := range items {
		sum.Count += it.Quantity
	}
	return sum, nil
}

func (s *CartStore) GetOrCreate(cartID string) error {
	_, err := s.db.Exec(
		"INSERT OR IGNORE INTO carts (cart_id) VALUES (?)",
		cartID,
	)
	return err
}

func (s *CartStore) AddItem(cartID string, productID int, maxStock int) error {
	_, err := s.db.Exec(`
                  INSERT INTO cart_items (cart_id, product_id, quantity)
                  VALUES (?, ?, 1)
                  ON CONFLICT(cart_id, product_id)
                  DO UPDATE SET quantity = MIN(quantity + 1, ?)`,
		cartID, productID, maxStock,
	)
	return err
}

func (s *CartStore) GetItems(cartID string) ([]CartItem, error) {
	rows, err := s.db.Query(
		"SELECT id, cart_id, product_id, quantity FROM cart_items WHERE cart_id = ?",
		cartID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CartItem
	for rows.Next() {
		var item CartItem
		if err := rows.Scan(&item.ID, &item.CartID, &item.ProductID, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *CartStore) GetItemDetails(cartID string) ([]CartItemDetail, error) {
	return itemDetails(s.db, cartID)
}

func (s *CartStore) TotalQuantity(cartID string) (int, error) {
	var total int
	err := s.db.QueryRow(
		"SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE cart_id = ?",
		cartID,
	).Scan(&total)
	return total, err
}

func (s *CartStore) UpdateQuantity(cartID string, productID int, qty int) error {
	_, err := s.db.Exec(
		"UPDATE cart_items SET quantity = ? WHERE cart_id = ? AND product_id = ?",
		qty, cartID, productID,
	)
	return err
}

func (s *CartStore) RemoveItem(cartID string, productID int) error {
	_, err := s.db.Exec(
		"DELETE FROM cart_items WHERE cart_id = ? AND product_id = ?",
		cartID, productID,
	)
	return err
}
