package order

import "github.com/tm-ox/go-datastar/internal/store/cart"

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

type OrderStore interface {
	Create(cartID string, items []cart.CartItemDetail) (int, error)
}
