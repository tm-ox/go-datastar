package cart

type CartItem struct {
	ID        int
	CartID    string
	ProductID int
	Quantity  int
}

type CartStore interface {
	GetOrCreate(cartID string) error
	AddItem(cartID string, productID int, maxStock int) error
	GetItems(cartID string) ([]CartItem, error)
	TotalQuantity(cartID string) (int, error)
	UpdateQuantity(cartID string, productID int, qty int) error
	RemoveItem(cartID string, productID int) error
	Clear(cartID string) error
}
