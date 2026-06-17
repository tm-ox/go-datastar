package cart

import "database/sql"

type SQLiteCartStore struct {
	db *sql.DB
}

func NewSQLiteCartStore(db *sql.DB) *SQLiteCartStore {
	return &SQLiteCartStore{db: db}
}

func (s *SQLiteCartStore) GetOrCreate(cartID string) error {
	_, err := s.db.Exec(
		"INSERT OR IGNORE INTO carts (cart_id) VALUES (?)",
		cartID,
	)
	return err
}

func (s *SQLiteCartStore) AddItem(cartID string, productID int, maxStock int) error {
	_, err := s.db.Exec(`
                  INSERT INTO cart_items (cart_id, product_id, quantity)
                  VALUES (?, ?, 1)
                  ON CONFLICT(cart_id, product_id)
                  DO UPDATE SET quantity = MIN(quantity + 1, ?)`,
		cartID, productID, maxStock,
	)
	return err
}

func (s *SQLiteCartStore) GetItems(cartID string) ([]CartItem, error) {
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

func (s *SQLiteCartStore) TotalQuantity(cartID string) (int, error) {
	var total int
	err := s.db.QueryRow(
		"SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE cart_id = ?",
		cartID,
	).Scan(&total)
	return total, err
}

func (s *SQLiteCartStore) UpdateQuantity(cartID string, productID int, qty int) error {
	_, err := s.db.Exec(
		"UPDATE cart_items SET quantity = ? WHERE cart_id = ? AND product_id = ?",
		qty, cartID, productID,
	)
	return err
}

func (s *SQLiteCartStore) RemoveItem(cartID string, productID int) error {
	_, err := s.db.Exec(
		"DELETE FROM cart_items WHERE cart_id = ? AND product_id = ?",
		cartID, productID,
	)
	return err
}

func (s *SQLiteCartStore) Clear(cartID string) error {
	_, err := s.db.Exec(
		"DELETE FROM cart_items WHERE cart_id = ?",
		cartID,
	)
	return err
}
